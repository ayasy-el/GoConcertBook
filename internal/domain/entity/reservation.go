package entity

import "time"

type Reservation struct {
	ID         string
	UserID     string
	EventID    string
	Category   string
	Qty        int
	Status     string
	ExpiredAt  time.Time
	CreatedAt  time.Time
}

const (
	ReservationStatusReserved  = "reserved"
	ReservationStatusConfirmed = "confirmed"
	ReservationStatusExpired   = "expired"
)
