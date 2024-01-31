package event

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Event struct {
	Id        string    `db:"id"`
	Name      string    `db:"name"`
	Capacity  int       `db:"capacity"`
	Start     time.Time `db:"start"`
	Location  string    `db:"location"`
	CreatedAt time.Time `db:"created_at"`
	EventResponse
}

type CreateEventRequest struct {
	Name           string `schema:"name"`
	Capacity       int    `schema:"capacity"`
	Start          string `schema:"start"`
	TimezoneOffset int    `schema:"timezoneOffset"`
	Location       string `schema:"location"`
}

type EventResponse struct {
	EventId   string    `db:"event_id"`
	UserId    string    `db:"user_id"`
	Going     sql.NullBool `db:"going"`
	UpdatedAt time.Time `db:"updated_at"`
}

type RespondEventRequest struct {
	Id    string `schema:"id"`
	Going bool   `schema:"going"`
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
	newId, err := gonanoid.New()
	if err != nil {
		return err
	}

	stmt := `
        INSERT INTO event (id, name, capacity, start, location, created_at)
        VALUES (?, ?, ?, ?, ?, ?)
    `
	args := []any{
		newId,
		e.Name,
		e.Capacity,
		e.Start,
		e.Location,
		e.CreatedAt,
	}

	_, err = s.db.Exec(stmt, args...)
	return err
}

func (s *Store) GetCurrent(userId string) ([]Event, error) {
	stmt := `
        SELECT e.id, e.name, e.capacity, e.start, e.location, e.created_at, er.going FROM event AS e
        LEFT JOIN event_response AS er ON e.id = er.event_id 
            AND er.user_id = ?
        WHERE datetime() <= datetime(start)
        ORDER BY start DESC
    `
	args := []any{userId}
	var events []Event

	err := s.db.Select(&events, stmt, args...)
	if err != nil {
		return []Event{}, err
	}

	return events, nil
}

func (s *Store) UpdateResponse(e EventResponse) error {
	stmt := `
        INSERT INTO event_response (event_id, user_id, going, updated_at)
        VALUES (?, ?, ?, ?)
        ON CONFLICT (event_id, user_id) DO UPDATE SET
            going = excluded.going,
            updated_at = excluded.updated_at
    `
	args := []any{
		e.EventId,
		e.UserId,
		e.Going,
		time.Now().UTC(),
	}

	_, err := s.db.Exec(stmt, args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetById(userId string, eventId string) (Event, error) {
	stmt := `
        SELECT e.id, e.name, e.capacity, e.start, e.location, e.created_at, er.going FROM event AS e
        LEFT JOIN event_response AS er ON e.id = er.event_id
            AND er.user_id = ?
        WHERE id = ?
    `
	args := []any{userId, eventId}

	var e Event
	err := s.db.Get(&e, stmt, args...)
	if err != nil {
		return Event{}, err
	}

	return e, nil
}
