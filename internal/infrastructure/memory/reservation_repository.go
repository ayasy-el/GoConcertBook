package memory

import (
	"sync"

	"concert-booking/internal/domain/entity"
)

type ReservationRepository struct {
	mu    sync.RWMutex
	items map[string]entity.Reservation
}

func NewReservationRepository() *ReservationRepository {
	return &ReservationRepository{items: map[string]entity.Reservation{}}
}

func (r *ReservationRepository) Upsert(reservation entity.Reservation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[reservation.ID] = reservation
	return nil
}

func (r *ReservationRepository) FindByID(id string) (entity.Reservation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.items[id]
	if !ok {
		return entity.Reservation{}, errMemoryNotFound
	}
	return v, nil
}

func (r *ReservationRepository) UpdateStatus(id, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.items[id]
	if !ok {
		return errMemoryNotFound
	}
	v.Status = status
	r.items[id] = v
	return nil
}
