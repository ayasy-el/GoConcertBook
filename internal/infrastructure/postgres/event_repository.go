package postgres

import (
	"database/sql"

	"concert-booking/internal/domain/entity"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(event entity.Event) error {
	_, err := r.db.Exec(`INSERT INTO events(id, name, date, created_at) VALUES ($1,$2,$3,$4)`, event.ID, event.Name, event.Date, event.CreatedAt)
	return err
}

func (r *EventRepository) FindByID(id string) (entity.Event, error) {
	var out entity.Event
	err := r.db.QueryRow(`SELECT id, name, date, created_at FROM events WHERE id=$1`, id).Scan(&out.ID, &out.Name, &out.Date, &out.CreatedAt)
	return out, err
}
