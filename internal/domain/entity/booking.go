package entity

import "time"

type Booking struct {
	ID            string
	ReservationID string
	PaymentStatus string
	CreatedAt     time.Time
}
