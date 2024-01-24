package event

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type Event struct {
	Name      string
	Limit     int
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
	return nil
}
