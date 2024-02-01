package event

import (
	"errors"
	"github/mattfan00/jvbe/user"
	"time"

	"github.com/jmoiron/sqlx"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Event struct {
	Id            string    `db:"id"`
	Name          string    `db:"name"`
	Capacity      int       `db:"capacity"`
	Start         time.Time `db:"start"`
	Location      string    `db:"location"`
	CreatedAt     time.Time `db:"created_at"`
	AttendeeCount int       `db:"attendee_count"`
	EventResponse
}

func (e Event) SpotsLeft() int {
	return e.Capacity - e.AttendeeCount
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
	Going     bool      `db:"going"`
	UpdatedAt time.Time `db:"updated_at"`
	user.User
}

type RespondEventRequest struct {
	Id    string `schema:"id"`
	Going bool   `schema:"going"`
}

type EventDetailed struct {
	Event
	EventResponses []EventResponse
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
        SELECT 
            e.id, e.name, e.capacity, e.start, e.location, e.created_at
            , COALESCE(er.going, false) AS going
            , COALESCE (ec.attendee_count, 0) AS attendee_count
        FROM event AS e
        LEFT JOIN event_response AS er ON e.id = er.event_id 
            AND er.user_id = ?
        LEFT JOIN (
            SELECT event_id, COUNT(*) AS attendee_count FROM event_response
            WHERE going = TRUE 
            GROUP BY event_id
        ) AS ec ON e.id = ec.event_id
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

func prepareGetById(eventId string, userId string) (string, []any) {
	stmt := `
        SELECT 
            e.id, e.name, e.capacity, e.start, e.location, e.created_at
            , COALESCE(er.going, false) AS going
            , (
                SELECT COUNT(*) FROM event_response
                WHERE event_id = ? AND going = TRUE 
            ) AS attendee_count
        FROM event AS e
        LEFT JOIN event_response AS er ON e.id = er.event_id
            AND er.user_id = ?
        WHERE id = ?
    `
	args := []any{eventId, userId, eventId}

	return stmt, args
}

func (s *Store) GetById(eventId string, userId string) (Event, error) {
	stmt, args := prepareGetById(eventId, userId)

	var event Event
	err := s.db.Get(&event, stmt, args...)
	if err != nil {
		return Event{}, err
	}

	return event, nil
}

func (s *Store) UpdateResponse(e EventResponse) (Event, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return Event{}, err
	}

	defer tx.Rollback()

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

	_, err = tx.Exec(stmt, args...)
	if err != nil {
		return Event{}, err
	}

	stmt, args = prepareGetById(e.EventId, e.UserId)

	var event Event
	err = tx.Get(&event, stmt, args...)
	if err != nil {
		return Event{}, err
	}
	// this statement along with the lock and transaction is the backbone of the concurrency issue
	if event.SpotsLeft() < 0 {
		return Event{}, errors.New("no spots left")
	}

	err = tx.Commit()
	if err != nil {
		return Event{}, err
	}

	return event, nil
}

func (s *Store) GetResponsesByEventId(eventId string) ([]EventResponse, error) {
	stmt := `
        SELECT er.event_id, er.user_id, er.going, u.full_name FROM event_response AS er
        INNER JOIN user AS u ON er.user_id = u.id
        WHERE er.event_id = ? AND er.going = TRUE
    `
	args := []any{eventId}

	var responses []EventResponse
	err := s.db.Select(&responses, stmt, args...)
	if err != nil {
		return []EventResponse{}, err
	}

	return responses, nil
}
