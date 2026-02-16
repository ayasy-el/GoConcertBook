package postgres

import (
	"database/sql"

	"concert-booking/internal/domain/entity"
)

type ReservationRepository struct {
	db *sql.DB
}

func NewReservationRepository(db *sql.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

func (r *ReservationRepository) Upsert(reservation entity.Reservation) error {
	_, err := r.db.Exec(`
	INSERT INTO reservations(id, user_id, event_id, category, qty, status, expired_at, created_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	ON CONFLICT (id) DO UPDATE SET
	status = EXCLUDED.status,
	expired_at = EXCLUDED.expired_at
	`, reservation.ID, reservation.UserID, reservation.EventID, reservation.Category, reservation.Qty, reservation.Status, reservation.ExpiredAt, reservation.CreatedAt)
	return err
}

func (r *ReservationRepository) FindByID(id string) (entity.Reservation, error) {
	var out entity.Reservation
	err := r.db.QueryRow(`SELECT id, user_id, event_id, category, qty, status, expired_at, created_at FROM reservations WHERE id=$1`, id).
		Scan(&out.ID, &out.UserID, &out.EventID, &out.Category, &out.Qty, &out.Status, &out.ExpiredAt, &out.CreatedAt)
	return out, err
}

func (r *ReservationRepository) UpdateStatus(id, status string) error {
	_, err := r.db.Exec(`UPDATE reservations SET status=$2 WHERE id=$1`, id, status)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`
	INSERT INTO reservations(id, user_id, event_id, category, qty, status, expired_at, created_at)
	VALUES ($1,'','', '',0,$2,NOW(),NOW())
	ON CONFLICT (id) DO NOTHING
	`, id, status)
	return err
}
