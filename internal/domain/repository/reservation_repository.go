package repository

import "concert-booking/internal/domain/entity"

type ReservationRepository interface {
	Upsert(reservation entity.Reservation) error
	FindByID(id string) (entity.Reservation, error)
	UpdateStatus(id, status string) error
}

type BookingRepository interface {
	CreateIfNotExists(booking entity.Booking) (bool, error)
	FindByReservationID(reservationID string) (entity.Booking, error)
}
