package memory

import (
	"errors"
	"sync"

	"concert-booking/internal/domain/entity"
)

var errMemoryNotFound = errors.New("not found")

type EventRepository struct {
	mu     sync.RWMutex
	events map[string]entity.Event
}

func NewEventRepository() *EventRepository {
	return &EventRepository{events: map[string]entity.Event{}}
}

func (r *EventRepository) Create(event entity.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events[event.ID] = event
	return nil
}

func (r *EventRepository) FindByID(id string) (entity.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.events[id]
	if !ok {
		return entity.Event{}, errMemoryNotFound
	}
	return e, nil
}
