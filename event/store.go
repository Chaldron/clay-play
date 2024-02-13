package event

import (
	"database/sql"
	"errors"
	"github/mattfan00/jvbe/user"
	"time"

	"github.com/jmoiron/sqlx"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Event struct {
	Id                 string    `db:"id"`
	Name               string    `db:"name"`
	Capacity           int       `db:"capacity"`
	Start              time.Time `db:"start"`
	Location           string    `db:"location"`
	CreatedAt          time.Time `db:"created_at"`
	Creator			   string    `db:"creator"`
	TotalAttendeeCount int       `db:"total_attendee_count"`
}

func (e Event) SpotsLeft() int {
	return e.Capacity - e.TotalAttendeeCount
}

type CreateEventRequest struct {
	Name           string `schema:"name"`
	Capacity       int    `schema:"capacity"`
	Start          string `schema:"start"`
	TimezoneOffset int    `schema:"timezoneOffset"`
	Location       string `schema:"location"`
}

type EventResponse struct {
	EventId       string    `db:"event_id"`
	UserId        string    `db:"user_id"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
	AttendeeCount int       `db:"attendee_count"`
	OnWaitlist    bool      `db:"on_waitlist"`
	user.User
}

func (e EventResponse) PlusOnes() int {
	return e.AttendeeCount - 1
}

type RespondEventRequest struct {
	Id            string `schema:"id"`
	AttendeeCount int    `schema:"attendee_count"`
}

type EventDetailed struct {
	Event
	UserResponse *EventResponse
	Responses    []EventResponse
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
        INSERT INTO event (id, name, capacity, start, location, created_at, creator)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `
	args := []any{
		newId,
		e.Name,
		e.Capacity,
		e.Start,
		e.Location,
		e.CreatedAt,
		e.Creator,
	}

	_, err = s.db.Exec(stmt, args...)
	return err
}

func (s *Store) GetCurrent(userId string) ([]Event, error) {
	stmt := `
        SELECT 
            e.id, e.name, e.capacity, e.start, e.location, e.created_at, e.creator
		    , COALESCE (ec.total_attendee_count, 0) AS total_attendee_count
        FROM event AS e
        LEFT JOIN (
            SELECT event_id, SUM(attendee_count) AS total_attendee_count FROM event_response
            WHERE on_waitlist = FALSE
            GROUP BY event_id
        ) AS ec ON e.id = ec.event_id
        WHERE datetime() <= datetime(start) AND is_deleted = FALSE
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

func (s *Store) GetById(eventId string) (Event, error) {
	stmt := `
        SELECT
            e.id, e.name, e.capacity, e.start, e.location, e.created_at, e.creator
            , COALESCE((
                SELECT SUM(attendee_count) FROM event_response
                WHERE event_id = ? AND on_waitlist = FALSE
            ), 0) AS total_attendee_count
        FROM event AS e
        WHERE id = ? AND is_deleted = FALSE 
    `
	args := []any{eventId, eventId}

	var event Event
	err := s.db.Get(&event, stmt, args...)
	if err != nil {
		return Event{}, err
	}

	return event, nil
}

func (s *Store) UpdateResponse(e EventResponse) error {
	stmt := `
        INSERT INTO event_response (event_id, user_id, created_at, updated_at, attendee_count, on_waitlist)
        VALUES (?, ?, ?, ?, ?, ?)
        ON CONFLICT (event_id, user_id) DO UPDATE SET
            updated_at = excluded.updated_at,
            attendee_count = excluded.attendee_count
    `

	now := time.Now().UTC()
	args := []any{
		e.EventId,
		e.UserId,
		now,
		now,
		e.AttendeeCount,
		e.OnWaitlist,
	}

	_, err := s.db.Exec(stmt, args...)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) DeleteResponse(eventId string, userId string) error {
	stmt := `
        DELETE FROM event_response
        WHERE event_id = ? AND user_id = ?
    `
	args := []any{eventId, userId}

	_, err := s.db.Exec(stmt, args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetResponsesByEventId(eventId string) ([]EventResponse, error) {
	stmt := `
        SELECT er.event_id, er.user_id, er.attendee_count, u.full_name, er.created_at, er.on_waitlist
        FROM event_response AS er
        INNER JOIN user AS u ON er.user_id = u.id
        WHERE er.event_id = ?
        ORDER BY er.created_at
    `
	args := []any{eventId}

	rows, err := s.db.Query(stmt, args...)
	if err != nil {
		return []EventResponse{}, err
	}
	defer rows.Close()
	var responses []EventResponse
	for rows.Next() {
		var i EventResponse
		if err := rows.Scan(&i.EventId, &i.UserId, &i.AttendeeCount, &i.FullName, &i.CreatedAt, &i.OnWaitlist); err != nil {
			return []EventResponse{}, err
		}
		responses = append(responses, i)
	}
	if err := rows.Err(); err != nil {
		return []EventResponse{}, err
	}

	return responses, nil
}

func (s *Store) GetWaitlist(eventId string, limit int) ([]EventResponse, error) {
	stmt := `
        SELECT er.event_id, er.user_id, er.on_waitlist
        FROM event_response AS er
        INNER JOIN user AS u ON er.user_id = u.id
        WHERE er.event_id = ? AND er.on_waitlist = TRUE
        ORDER BY er.created_at
        LIMIT ?
    `
	args := []any{eventId, limit}

	var waitlist []EventResponse
	err := s.db.Select(&waitlist, stmt, args...)
	if err != nil {
		return []EventResponse{}, err
	}

	return waitlist, nil
}

func (s *Store) UpdateWaitlist(reqs []EventResponse) error {
	t, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer t.Rollback()

	stmt := `
        UPDATE event_response
        SET on_waitlist = 0
        WHERE event_id = ? AND user_id = ?
    `

	for _, req := range reqs {
		args := []any{req.EventId, req.UserId}

		_, err := t.Exec(stmt, args...)
		if err != nil {
			return err
		}
	}

	err = t.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetUserResponse(eventId string, userId string) (*EventResponse, error) {
	stmt := `
        SELECT event_id, attendee_count, on_waitlist
        FROM event_response
        WHERE event_id = ? AND user_id = ?
    `
	args := []any{eventId, userId}

	var response EventResponse
	err := s.db.Get(&response, stmt, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &response, nil
}

func (s *Store) DeleteById(id string) error {
	stmt := `
        UPDATE event
        SET is_deleted = TRUE
        WHERE id = ?
    `
	args := []any{id}

	_, err := s.db.Exec(stmt, args...)
	if err != nil {
		return err
	}

	return nil
}
