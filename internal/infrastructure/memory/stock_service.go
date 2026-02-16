package memory

import (
	"context"
	"sync"
	"time"

	"concert-booking/internal/domain/service"
)

type stockKey struct {
	eventID  string
	category string
}

type StockService struct {
	mu           sync.Mutex
	stocks       map[stockKey]int
	reservations map[string]service.ReservationMeta
}

func NewStockService() *StockService {
	return &StockService{stocks: map[stockKey]int{}, reservations: map[string]service.ReservationMeta{}}
}

func (s *StockService) InitStock(_ context.Context, eventID, category string, total int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stocks[stockKey{eventID: eventID, category: category}] = total
	return nil
}

func (s *StockService) Reserve(_ context.Context, meta service.ReservationMeta, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := stockKey{eventID: meta.EventID, category: meta.Category}
	stock := s.stocks[k]
	if stock < meta.Qty {
		return service.ErrOutOfStock
	}
	s.stocks[k] = stock - meta.Qty
	meta.Status = "reserved"
	meta.ExpiredAt = time.Now().Add(ttl)
	s.reservations[meta.ReservationID] = meta
	return nil
}

func (s *StockService) GetReservation(_ context.Context, reservationID string) (service.ReservationMeta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.reservations[reservationID]
	if !ok {
		return service.ReservationMeta{}, service.ErrReservationNotFound
	}
	if time.Now().After(v.ExpiredAt) && v.Status == "reserved" {
		k := stockKey{eventID: v.EventID, category: v.Category}
		s.stocks[k] += v.Qty
		v.Status = "expired"
		s.reservations[reservationID] = v
		return service.ReservationMeta{}, service.ErrReservationNotFound
	}
	return v, nil
}

func (s *StockService) ConfirmReservation(_ context.Context, reservationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.reservations[reservationID]
	if !ok {
		return service.ErrReservationNotFound
	}
	if v.Status != "reserved" {
		return service.ErrReservationFinalized
	}
	v.Status = "confirmed"
	s.reservations[reservationID] = v
	return nil
}

func (s *StockService) ReleaseReservation(_ context.Context, reservationID string) (service.ReservationMeta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.reservations[reservationID]
	if !ok {
		return service.ReservationMeta{}, service.ErrReservationNotFound
	}
	if v.Status != "reserved" {
		return service.ReservationMeta{}, service.ErrReservationFinalized
	}
	k := stockKey{eventID: v.EventID, category: v.Category}
	s.stocks[k] += v.Qty
	v.Status = "expired"
	s.reservations[reservationID] = v
	return v, nil
}

func (s *StockService) ReleaseExpired(_ context.Context, now time.Time, limit int) ([]service.ReservationMeta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]service.ReservationMeta, 0, limit)
	for id, v := range s.reservations {
		if len(out) >= limit {
			break
		}
		if v.Status == "reserved" && !v.ExpiredAt.After(now) {
			k := stockKey{eventID: v.EventID, category: v.Category}
			s.stocks[k] += v.Qty
			v.Status = "expired"
			s.reservations[id] = v
			out = append(out, v)
		}
	}
	return out, nil
}
