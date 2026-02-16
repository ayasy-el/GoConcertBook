package memory

import (
	"errors"
	"sync"

	"concert-booking/internal/domain/entity"
)

var errCategoryAlreadyExists = errors.New("category already exists")

type TicketCategoryRepository struct {
	mu         sync.RWMutex
	byEvent    map[string][]entity.TicketCategory
	byEventKey map[string]entity.TicketCategory
}

func NewTicketCategoryRepository() *TicketCategoryRepository {
	return &TicketCategoryRepository{
		byEvent:    map[string][]entity.TicketCategory{},
		byEventKey: map[string]entity.TicketCategory{},
	}
}

func (r *TicketCategoryRepository) Create(category entity.TicketCategory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := category.EventID + ":" + category.Name
	if _, ok := r.byEventKey[k]; ok {
		return errCategoryAlreadyExists
	}
	r.byEventKey[k] = category
	r.byEvent[category.EventID] = append(r.byEvent[category.EventID], category)
	return nil
}

func (r *TicketCategoryRepository) FindByEventID(eventID string) ([]entity.TicketCategory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := r.byEvent[eventID]
	out := make([]entity.TicketCategory, len(items))
	copy(out, items)
	return out, nil
}

func (r *TicketCategoryRepository) FindByEventAndName(eventID, name string) (entity.TicketCategory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	k := eventID + ":" + name
	c, ok := r.byEventKey[k]
	if !ok {
		return entity.TicketCategory{}, errMemoryNotFound
	}
	return c, nil
}
