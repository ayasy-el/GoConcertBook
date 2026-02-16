package postgres

import (
	"database/sql"

	"concert-booking/internal/domain/entity"
)

type TicketCategoryRepository struct {
	db *sql.DB
}

func NewTicketCategoryRepository(db *sql.DB) *TicketCategoryRepository {
	return &TicketCategoryRepository{db: db}
}

func (r *TicketCategoryRepository) Create(category entity.TicketCategory) error {
	_, err := r.db.Exec(`INSERT INTO ticket_categories(id, event_id, name, total_stock, price) VALUES ($1,$2,$3,$4,$5)`, category.ID, category.EventID, category.Name, category.TotalStock, category.Price)
	return err
}

func (r *TicketCategoryRepository) FindByEventID(eventID string) ([]entity.TicketCategory, error) {
	rows, err := r.db.Query(`SELECT id, event_id, name, total_stock, price FROM ticket_categories WHERE event_id=$1`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]entity.TicketCategory, 0)
	for rows.Next() {
		var c entity.TicketCategory
		if err := rows.Scan(&c.ID, &c.EventID, &c.Name, &c.TotalStock, &c.Price); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *TicketCategoryRepository) FindByEventAndName(eventID, name string) (entity.TicketCategory, error) {
	var c entity.TicketCategory
	err := r.db.QueryRow(`SELECT id, event_id, name, total_stock, price FROM ticket_categories WHERE event_id=$1 AND name=$2`, eventID, name).Scan(&c.ID, &c.EventID, &c.Name, &c.TotalStock, &c.Price)
	return c, err
}
