package repository

import "concert-booking/internal/domain/entity"

type EventRepository interface {
	Create(event entity.Event) error
	FindByID(id string) (entity.Event, error)
}
