package event

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type Event struct {
	Id        int       `db:"id"`
	Name      string    `db:"name"`
	Capacity  int       `db:"capacity"`
	Start     time.Time `db:"start"`
	Location  string    `db:"location"`
	CreatedAt time.Time `db:"created_at"`
}

type Store struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) InsertOne(e Event) error {
	stmt := `
        INSERT INTO event (name, capacity, start, location, created_at)
        VALUES (?, ?, ?, ?, ?)
    `
	args := []any{
        e.Name,
		e.Capacity,
		e.Start,
		e.Location,
		e.CreatedAt,
	}

	_, err := s.db.Exec(stmt, args...)
	return err
}

func (s *Store) GetCurrent() ([]Event, error) {
	stmt := `
        SELECT id, name, capacity, start, location, created_at FROM event
        WHERE datetime() <= datetime(start)
        ORDER BY start DESC
    `
	var events []Event

	err := s.db.Select(&events, stmt)
	if err != nil {
		return []Event{}, err
	}

	return events, nil
}
