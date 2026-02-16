package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync/atomic"
	"time"

	"concert-booking/internal/domain/entity"
	"concert-booking/internal/domain/repository"
	"concert-booking/internal/domain/service"
)

var ErrQueueFull = errors.New("queue is full")

type ReservationUsecase struct {
	categories      repository.TicketCategoryRepository
	reservations    repository.ReservationRepository
	bookings        repository.BookingRepository
	stock           service.StockService
	producer        service.EventProducer
	now             func() time.Time
	newID           func() string
	ttl             time.Duration
	queueThreshold  int64
	waitingRequests atomic.Int64
	gate            chan struct{}
	persistSync     bool
}

func NewReservationUsecase(categories repository.TicketCategoryRepository, reservations repository.ReservationRepository, bookings repository.BookingRepository, stock service.StockService, producer service.EventProducer, now func() time.Time, newID func() string, ttl time.Duration, queueThreshold, workerPoolSize int, persistSync bool) *ReservationUsecase {
	if workerPoolSize <= 0 {
		workerPoolSize = 1
	}
	return &ReservationUsecase{
		categories:     categories,
		reservations:   reservations,
		bookings:       bookings,
		stock:          stock,
		producer:       producer,
		now:            now,
		newID:          newID,
		ttl:            ttl,
		queueThreshold: int64(queueThreshold),
		gate:           make(chan struct{}, workerPoolSize),
		persistSync:    persistSync,
	}
}

func (u *ReservationUsecase) Reserve(ctx context.Context, userID, eventID, category string, qty int) (entity.Reservation, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(eventID) == "" || strings.TrimSpace(category) == "" || qty <= 0 {
		return entity.Reservation{}, ErrInvalidInput
	}
	if u.waitingRequests.Add(1) > u.queueThreshold {
		u.waitingRequests.Add(-1)
		return entity.Reservation{}, ErrQueueFull
	}
	defer u.waitingRequests.Add(-1)

	select {
	case u.gate <- struct{}{}:
		defer func() { <-u.gate }()
	case <-ctx.Done():
		return entity.Reservation{}, ctx.Err()
	}

	category = strings.ToUpper(strings.TrimSpace(category))
	if _, err := u.categories.FindByEventAndName(eventID, category); err != nil {
		return entity.Reservation{}, ErrNotFound
	}

	res := entity.Reservation{
		ID:        u.newID(),
		UserID:    userID,
		EventID:   eventID,
		Category:  category,
		Qty:       qty,
		Status:    entity.ReservationStatusReserved,
		ExpiredAt: u.now().Add(u.ttl),
		CreatedAt: u.now(),
	}

	meta := service.ReservationMeta{
		ReservationID: res.ID,
		UserID:        userID,
		EventID:       eventID,
		Category:      category,
		Qty:           qty,
		Status:        entity.ReservationStatusReserved,
		ExpiredAt:     res.ExpiredAt,
	}
	if err := u.stock.Reserve(ctx, meta, u.ttl); err != nil {
		if errors.Is(err, service.ErrOutOfStock) {
			return entity.Reservation{}, service.ErrOutOfStock
		}
		return entity.Reservation{}, err
	}

	if u.persistSync {
		if err := u.reservations.Upsert(res); err != nil {
			return entity.Reservation{}, err
		}
	}

	payload, _ := json.Marshal(res)
	if err := u.producer.Publish(ctx, "ticket.reserved", eventID, payload); err != nil {
		return entity.Reservation{}, err
	}
	return res, nil
}

func (u *ReservationUsecase) StartExpiryReaper(ctx context.Context, interval time.Duration, batch int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			_ = u.ReleaseExpired(ctx, now, batch)
		}
	}
}

func (u *ReservationUsecase) Confirm(ctx context.Context, reservationID string, paymentOK bool) (entity.Booking, error) {
	if strings.TrimSpace(reservationID) == "" {
		return entity.Booking{}, ErrInvalidInput
	}
	resMeta, err := u.stock.GetReservation(ctx, reservationID)
	if err != nil {
		if errors.Is(err, service.ErrReservationNotFound) {
			if existing, ferr := u.bookings.FindByReservationID(reservationID); ferr == nil {
				return existing, nil
			}
			return entity.Booking{}, ErrNotFound
		}
		return entity.Booking{}, err
	}

	if !paymentOK {
		_, _ = u.stock.ReleaseReservation(ctx, reservationID)
		_ = u.reservations.UpdateStatus(reservationID, entity.ReservationStatusExpired)
		payload, _ := json.Marshal(map[string]string{"reservation_id": reservationID, "status": "expired"})
		_ = u.producer.Publish(ctx, "ticket.expired", resMeta.EventID, payload)
		return entity.Booking{}, errors.New("payment failed")
	}

	if err := u.stock.ConfirmReservation(ctx, reservationID); err != nil {
		if errors.Is(err, service.ErrReservationFinalized) {
			if existing, ferr := u.bookings.FindByReservationID(reservationID); ferr == nil {
				return existing, nil
			}
		}
		return entity.Booking{}, err
	}
	_ = u.reservations.UpdateStatus(reservationID, entity.ReservationStatusConfirmed)

	booking := entity.Booking{
		ID:            u.newID(),
		ReservationID: reservationID,
		PaymentStatus: "paid",
		CreatedAt:     u.now(),
	}
	created, err := u.bookings.CreateIfNotExists(booking)
	if err != nil {
		return entity.Booking{}, err
	}
	if !created {
		existing, err := u.bookings.FindByReservationID(reservationID)
		if err != nil {
			return entity.Booking{}, err
		}
		return existing, nil
	}

	payload, _ := json.Marshal(booking)
	if err := u.producer.Publish(ctx, "ticket.confirmed", resMeta.EventID, payload); err != nil {
		return entity.Booking{}, err
	}
	return booking, nil
}

func (u *ReservationUsecase) ReleaseExpired(ctx context.Context, now time.Time, batch int) error {
	items, err := u.stock.ReleaseExpired(ctx, now, batch)
	if err != nil {
		return err
	}
	for _, item := range items {
		_ = u.reservations.UpdateStatus(item.ReservationID, entity.ReservationStatusExpired)
		payload, _ := json.Marshal(map[string]string{"reservation_id": item.ReservationID, "status": "expired"})
		_ = u.producer.Publish(ctx, "ticket.expired", item.EventID, payload)
	}
	return nil
}
