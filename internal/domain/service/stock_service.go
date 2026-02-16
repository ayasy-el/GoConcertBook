package service

import (
	"context"
	"errors"
	"time"
)

var (
	ErrOutOfStock           = errors.New("out of stock")
	ErrReservationNotFound  = errors.New("reservation not found")
	ErrReservationFinalized = errors.New("reservation already finalized")
)

type ReservationMeta struct {
	ReservationID string
	UserID        string
	EventID       string
	Category      string
	Qty           int
	Status        string
	ExpiredAt     time.Time
}

type StockService interface {
	InitStock(ctx context.Context, eventID, category string, total int) error
	GetStocks(ctx context.Context, eventID string, categories []string) (map[string]int, error)
	Reserve(ctx context.Context, meta ReservationMeta, ttl time.Duration) error
	GetReservation(ctx context.Context, reservationID string) (ReservationMeta, error)
	ConfirmReservation(ctx context.Context, reservationID string) error
	ReleaseReservation(ctx context.Context, reservationID string) (ReservationMeta, error)
	ReleaseExpired(ctx context.Context, now time.Time, limit int) ([]ReservationMeta, error)
}

type EventProducer interface {
	Publish(ctx context.Context, topic string, key string, value []byte) error
}
