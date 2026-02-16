package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"concert-booking/internal/domain/service"

	goredis "github.com/redis/go-redis/v9"
)

type StockService struct {
	client *goredis.Client
}

func NewStockService(addr, password string) *StockService {
	return &StockService{client: goredis.NewClient(&goredis.Options{Addr: addr, Password: password})}
}

func (s *StockService) Client() *goredis.Client {
	return s.client
}

func (s *StockService) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

func (s *StockService) InitStock(ctx context.Context, eventID, category string, total int) error {
	return s.client.SetNX(ctx, stockKey(eventID, category), total, 0).Err()
}

func (s *StockService) GetStocks(ctx context.Context, eventID string, categories []string) (map[string]int, error) {
	if len(categories) == 0 {
		return map[string]int{}, nil
	}
	keys := make([]string, 0, len(categories))
	for _, category := range categories {
		keys = append(keys, stockKey(eventID, category))
	}
	values, err := s.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	out := make(map[string]int, len(categories))
	for i, category := range categories {
		switch v := values[i].(type) {
		case string:
			n, _ := strconv.Atoi(v)
			out[category] = n
		case int64:
			out[category] = int(v)
		case nil:
			out[category] = 0
		default:
			n, _ := strconv.Atoi(fmt.Sprint(v))
			out[category] = n
		}
	}
	return out, nil
}

func (s *StockService) Reserve(ctx context.Context, meta service.ReservationMeta, ttl time.Duration) error {
	payload, _ := json.Marshal(meta)
	expAt := strconv.FormatInt(meta.ExpiredAt.Unix(), 10)
	ttlSec := strconv.FormatInt(int64(ttl/time.Second), 10)
	res, err := s.client.Eval(ctx, `
local stock = tonumber(redis.call('GET', KEYS[1]) or '0')
local qty = tonumber(ARGV[1])
if stock < qty then
  return 0
end
redis.call('DECRBY', KEYS[1], qty)
redis.call('SET', KEYS[2], ARGV[2], 'EX', ARGV[3])
redis.call('HSET', KEYS[3], 'event_id', ARGV[4], 'category', ARGV[5], 'qty', ARGV[1], 'user_id', ARGV[6], 'status', 'reserved', 'expired_at', ARGV[7])
redis.call('EXPIRE', KEYS[3], 86400)
redis.call('ZADD', KEYS[4], ARGV[7], ARGV[8])
return 1
`, []string{stockKey(meta.EventID, meta.Category), reservationKey(meta.ReservationID), reservationMetaKey(meta.ReservationID), expirySetKey()},
		meta.Qty, string(payload), ttlSec, meta.EventID, meta.Category, meta.UserID, expAt, meta.ReservationID).Int()
	if err != nil {
		return err
	}
	if res != 1 {
		return service.ErrOutOfStock
	}
	return nil
}

func (s *StockService) GetReservation(ctx context.Context, reservationID string) (service.ReservationMeta, error) {
	metaMap, err := s.client.HGetAll(ctx, reservationMetaKey(reservationID)).Result()
	if err != nil {
		return service.ReservationMeta{}, err
	}
	if len(metaMap) == 0 {
		return service.ReservationMeta{}, service.ErrReservationNotFound
	}
	status := metaMap["status"]
	if status != "reserved" && status != "confirmed" {
		return service.ReservationMeta{}, service.ErrReservationNotFound
	}
	expUnix, _ := strconv.ParseInt(metaMap["expired_at"], 10, 64)
	meta := service.ReservationMeta{
		ReservationID: reservationID,
		UserID:        metaMap["user_id"],
		EventID:       metaMap["event_id"],
		Category:      metaMap["category"],
		Status:        status,
		ExpiredAt:     time.Unix(expUnix, 0),
	}
	meta.Qty, _ = strconv.Atoi(metaMap["qty"])
	if time.Now().After(meta.ExpiredAt) && meta.Status == "reserved" {
		return service.ReservationMeta{}, service.ErrReservationNotFound
	}
	return meta, nil
}

func (s *StockService) ConfirmReservation(ctx context.Context, reservationID string) error {
	res, err := s.client.Eval(ctx, `
local status = redis.call('HGET', KEYS[1], 'status')
if not status then
  return -1
end
if status ~= 'reserved' then
  return 0
end
redis.call('HSET', KEYS[1], 'status', 'confirmed')
redis.call('DEL', KEYS[2])
redis.call('ZREM', KEYS[3], ARGV[1])
return 1
`, []string{reservationMetaKey(reservationID), reservationKey(reservationID), expirySetKey()}, reservationID).Int()
	if err != nil {
		return err
	}
	switch res {
	case -1:
		return service.ErrReservationNotFound
	case 0:
		return service.ErrReservationFinalized
	}
	return nil
}

func (s *StockService) ReleaseReservation(ctx context.Context, reservationID string) (service.ReservationMeta, error) {
	meta, err := s.GetReservation(ctx, reservationID)
	if err != nil {
		return service.ReservationMeta{}, err
	}
	res, err := s.client.Eval(ctx, `
local status = redis.call('HGET', KEYS[1], 'status')
if not status then
  return -1
end
if status ~= 'reserved' then
  return 0
end
redis.call('HSET', KEYS[1], 'status', 'expired')
redis.call('INCRBY', KEYS[2], ARGV[1])
redis.call('DEL', KEYS[3])
redis.call('ZREM', KEYS[4], ARGV[2])
return 1
`, []string{reservationMetaKey(reservationID), stockKey(meta.EventID, meta.Category), reservationKey(reservationID), expirySetKey()}, meta.Qty, reservationID).Int()
	if err != nil {
		return service.ReservationMeta{}, err
	}
	if res != 1 {
		return service.ReservationMeta{}, service.ErrReservationFinalized
	}
	meta.Status = "expired"
	return meta, nil
}

func (s *StockService) ReleaseExpired(ctx context.Context, now time.Time, limit int) ([]service.ReservationMeta, error) {
	ids, err := s.client.ZRangeByScore(ctx, expirySetKey(), &goredis.ZRangeBy{Min: "-inf", Max: strconv.FormatInt(now.Unix(), 10), Offset: 0, Count: int64(limit)}).Result()
	if err != nil {
		return nil, err
	}
	out := make([]service.ReservationMeta, 0, len(ids))
	for _, id := range ids {
		metaMap, err := s.client.HGetAll(ctx, reservationMetaKey(id)).Result()
		if err != nil || len(metaMap) == 0 {
			continue
		}
		if metaMap["status"] != "reserved" {
			_ = s.client.ZRem(ctx, expirySetKey(), id).Err()
			continue
		}
		qty, _ := strconv.Atoi(metaMap["qty"])
		expUnix, _ := strconv.ParseInt(metaMap["expired_at"], 10, 64)
		meta := service.ReservationMeta{ReservationID: id, UserID: metaMap["user_id"], EventID: metaMap["event_id"], Category: metaMap["category"], Qty: qty, Status: "reserved", ExpiredAt: time.Unix(expUnix, 0)}
		if _, err := s.ReleaseReservation(ctx, id); err != nil {
			if !errors.Is(err, service.ErrReservationFinalized) {
				continue
			}
		}
		meta.Status = "expired"
		out = append(out, meta)
	}
	return out, nil
}

func stockKey(eventID, category string) string { return fmt.Sprintf("stock:%s:%s", eventID, category) }
func reservationKey(id string) string          { return "reservation:" + id }
func reservationMetaKey(id string) string      { return "reservation_meta:" + id }
func expirySetKey() string                     { return "reservation_expiries" }
