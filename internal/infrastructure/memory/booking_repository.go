package memory

import (
	"sync"

	"concert-booking/internal/domain/entity"
)

type BookingRepository struct {
	mu              sync.RWMutex
	items           map[string]entity.Booking
	byReservationID map[string]string
}

func NewBookingRepository() *BookingRepository {
	return &BookingRepository{items: map[string]entity.Booking{}, byReservationID: map[string]string{}}
}

func (r *BookingRepository) CreateIfNotExists(booking entity.Booking) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byReservationID[booking.ReservationID]; ok {
		return false, nil
	}
	r.items[booking.ID] = booking
	r.byReservationID[booking.ReservationID] = booking.ID
	return true, nil
}

func (r *BookingRepository) FindByReservationID(reservationID string) (entity.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byReservationID[reservationID]
	if !ok {
		return entity.Booking{}, errMemoryNotFound
	}
	return r.items[id], nil
}
