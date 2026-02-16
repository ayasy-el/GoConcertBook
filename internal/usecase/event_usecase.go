package usecase

import (
	"errors"
	"strings"
	"time"

	"concert-booking/internal/domain/entity"
	"concert-booking/internal/domain/repository"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("not found")
)

type EventUsecase struct {
	events    repository.EventRepository
	categories repository.TicketCategoryRepository
	now       func() time.Time
	newID     func() string
}

func NewEventUsecase(events repository.EventRepository, categories repository.TicketCategoryRepository, now func() time.Time, newID func() string) *EventUsecase {
	return &EventUsecase{events: events, categories: categories, now: now, newID: newID}
}

func (u *EventUsecase) CreateEvent(name string, date time.Time) (entity.Event, error) {
	if strings.TrimSpace(name) == "" || date.IsZero() {
		return entity.Event{}, ErrInvalidInput
	}
	e := entity.Event{
		ID:        u.newID(),
		Name:      strings.TrimSpace(name),
		Date:      date.UTC(),
		CreatedAt: u.now().UTC(),
	}
	if err := u.events.Create(e); err != nil {
		return entity.Event{}, err
	}
	return e, nil
}

func (u *EventUsecase) CreateCategory(eventID, name string, totalStock int, price int64) (entity.TicketCategory, error) {
	if strings.TrimSpace(eventID) == "" || strings.TrimSpace(name) == "" || totalStock <= 0 || price < 0 {
		return entity.TicketCategory{}, ErrInvalidInput
	}
	if _, err := u.events.FindByID(eventID); err != nil {
		return entity.TicketCategory{}, ErrNotFound
	}
	c := entity.TicketCategory{
		ID:         u.newID(),
		EventID:    eventID,
		Name:       strings.ToUpper(strings.TrimSpace(name)),
		TotalStock: totalStock,
		Price:      price,
	}
	if err := u.categories.Create(c); err != nil {
		return entity.TicketCategory{}, err
	}
	return c, nil
}

func (u *EventUsecase) Availability(eventID string) (map[string]int, error) {
	if strings.TrimSpace(eventID) == "" {
		return nil, ErrInvalidInput
	}
	if _, err := u.events.FindByID(eventID); err != nil {
		return nil, ErrNotFound
	}
	categories, err := u.categories.FindByEventID(eventID)
	if err != nil {
		return nil, err
	}
	out := make(map[string]int, len(categories))
	for _, c := range categories {
		out[strings.ToLower(c.Name)] = c.TotalStock
	}
	return out, nil
}
