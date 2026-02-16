package repository

import "concert-booking/internal/domain/entity"

type TicketCategoryRepository interface {
	Create(category entity.TicketCategory) error
	FindByEventID(eventID string) ([]entity.TicketCategory, error)
	FindByEventAndName(eventID, name string) (entity.TicketCategory, error)
}
