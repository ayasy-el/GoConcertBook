package postgres

import (
	"database/sql"

	"concert-booking/internal/domain/entity"
)

type BookingRepository struct {
	db *sql.DB
}

func NewBookingRepository(db *sql.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) CreateIfNotExists(booking entity.Booking) (bool, error) {
	row := r.db.QueryRow(`
	INSERT INTO bookings(id, reservation_id, payment_status, created_at)
	VALUES ($1,$2,$3,$4)
	ON CONFLICT (reservation_id) DO NOTHING
	RETURNING id
	`, booking.ID, booking.ReservationID, booking.PaymentStatus, booking.CreatedAt)
	var id string
	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *BookingRepository) FindByReservationID(reservationID string) (entity.Booking, error) {
	var b entity.Booking
	err := r.db.QueryRow(`SELECT id, reservation_id, payment_status, created_at FROM bookings WHERE reservation_id=$1`, reservationID).
		Scan(&b.ID, &b.ReservationID, &b.PaymentStatus, &b.CreatedAt)
	return b, err
}
