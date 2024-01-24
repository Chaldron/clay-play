package event

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type Event struct {
	Name      string
	Capacity  int
	Start     time.Time
	Location  string
	CreatedAt time.Time
}

type Store struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) Insert(e Event) error {
	stmt := `
        INSERT INTO event (capacity, start, location, created_at)
        VALUES (?, ?, ?, ?)
    `
	args := []any{
		e.Capacity,
		e.Start,
		e.Location,
		e.CreatedAt,
	}

	_, err := s.db.Exec(stmt, args...)
	return err
}
