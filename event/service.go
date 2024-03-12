package event

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/mattfan00/jvbe/db"
)

type service struct {
	db *db.DB
}

func NewService(db *db.DB) *service {
	return &service{
		db: db,
	}
}

func serviceLog(format string, s ...any) {
	log.Printf("event/service.go: %s", fmt.Sprintf(format, s...))
}

func (s *service) Get(id string) (Event, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return Event{}, err
	}
	defer tx.Rollback()

	e, err := get(tx, id)
	return e, err
}

func (s *service) GetDetailed(id string, userId string) (EventDetailed, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return EventDetailed{}, err
	}
	defer tx.Rollback()

	e, err := get(tx, id)
	if err != nil {
		return EventDetailed{}, err
	}

	r, err := listResponses(tx, id)
	if err != nil {
		return EventDetailed{}, err
	}

	ur, err := getUserResponse(tx, id, userId)
	if err != nil {
		return EventDetailed{}, err
	}

	ed := EventDetailed{
		Event:        e,
		Responses:    r,
		UserResponse: ur,
	}

	return ed, nil
}

type ListFilter struct {
	UserId      string
	Upcoming    bool
	Past        bool
	Limit       int
	Offset      int
	OrderByDesc bool
}

func (s *service) List(f ListFilter) (EventList, error) {
	where, wargs := []string{}, []any{}

	where = append(where, "is_deleted = FALSE")
	if f.Upcoming {
		where = append(where, "datetime() <= datetime(start)")
	}
	if f.Past {
		where = append(where, "datetime() > datetime(start)")
	}

	// move the logic for determining if user can access event based off group from group service over to here
	if f.UserId != "" {
		where = append(where, "(e.group_id IS NULL OR e.group_id IN (SELECT group_id FROM user_group_member WHERE user_id = ?))")
		wargs = append(wargs, f.UserId)
	}

	orderByDir := "ASC"
	if f.OrderByDesc == true {
		orderByDir = "DESC"
	}

	stmt := `
        SELECT 
            e.id, e.name, e.capacity, e.start, e.location, e.created_at, e.creator_id
		    , COALESCE (ec.total_attendee_count, 0) AS total_attendee_count
            , e.group_id
        FROM event AS e
        LEFT JOIN (
            SELECT event_id, SUM(attendee_count) AS total_attendee_count FROM event_response
            WHERE on_waitlist = FALSE
            GROUP BY event_id
        ) AS ec ON e.id = ec.event_id
        WHERE ` + strings.Join(where, " AND ") + `
        ORDER BY start ` + orderByDir + `
        ` + db.FormatLimitOffset(f.Limit, f.Offset)
	args := []any{}
	args = append(args, wargs...)

	var events []Event
	err := s.db.Select(&events, stmt, args...)
	if err != nil {
		return EventList{}, err
	}

	return EventList{
		Events: events,
	}, nil
}

type CreateParams struct {
	Name      string
	GroupId   string
	Capacity  int
	Start     time.Time
	Location  string
	CreatorId string
}

func (s *service) Create(p CreateParams) (string, error) {
	newId, err := gonanoid.New()
	if err != nil {
		return "", err
	}

	stmt := `
        INSERT INTO event (id, name, group_id, capacity, start, location, created_at, creator_id)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `
	args := []any{
		newId,
		p.Name,
		sql.NullString{
			String: p.GroupId,
			Valid:  p.GroupId != "",
		},
		p.Capacity,
		p.Start,
		p.Location,
		time.Now().UTC(),
		p.CreatorId,
	}
	serviceLog("Create args %v", args)

	_, err = s.db.Exec(stmt, args...)
	return newId, err
}

type UpdateParams struct {
	Id       string
	Name     string
	Capacity int
	Start    time.Time
	Location string
}

func (s *service) Update(p UpdateParams) error {
	stmt := `
        UPDATE event
        SET name = ?, capacity = ?, start = ?, location = ?
        WHERE id = ?
    `
	args := []any{
		p.Name,
		p.Capacity,
		p.Start,
		p.Location,
		p.Id,
	}
	serviceLog("Update args %v", args)

	_, err := s.db.Exec(stmt, args...)
	return err
}

func (s *service) Delete(id string) error {
	stmt := `
        UPDATE event
        SET is_deleted = TRUE
        WHERE id = ?
    `
	args := []any{id}
	serviceLog("Delete args %v", args)

	_, err := s.db.Exec(stmt, args...)
	if err != nil {
		return err
	}

	return nil
}

type HandleResponseParams struct {
	UserId        string
	Id            string
	AttendeeCount int
}

func (s *service) HandleResponse(p HandleResponseParams) error {
	if p.AttendeeCount < 0 {
		return errors.New("cannot have less than 0 attendees")
	}

	if p.AttendeeCount > MaxAttendeeCount {
		return fmt.Errorf("maximum of %d plus one(s) allowed", MaxAttendeeCount-1)
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	e, err := get(tx, p.Id)
	if err != nil {
		return err
	}

	if e.IsPast {
		return errors.New("cannot respond to past events")
	}

	existingResponse, err := getUserResponse(tx, p.Id, p.UserId)
	if err != nil {
		return err
	}

	attendeeCountDelta := p.AttendeeCount
	if existingResponse != nil { // if a response exists already, need to factor the attendees in that one
		attendeeCountDelta -= existingResponse.AttendeeCount
	}

	if p.AttendeeCount == 0 { // just delete the response, I don't think it really matters to keep it in DB
		err := deleteResponse(tx, p.Id, p.UserId)
		if err != nil {
			return err
		}
	} else {
		// if theres no space for the response coming in, add the response to the waitlist
		// waitlist responses should ALWAYS be 1 attendee (no plus ones)
		addToWaitlist := e.SpotsLeft()-attendeeCountDelta < 0
		if addToWaitlist && p.AttendeeCount > 1 {
			return errors.New("no plus ones when adding to waitlist")
		}

		err = updateResponse(tx, updateResponseParams{
			EventId:       p.Id,
			UserId:        p.UserId,
			AttendeeCount: p.AttendeeCount,
			OnWaitlist:    addToWaitlist,
		})
		if err != nil {
			return err
		}
	}

	fromAttendee := !(existingResponse != nil && existingResponse.OnWaitlist)
	serviceLog("spots left:%d delta:%d attendee:%t", e.SpotsLeft(), attendeeCountDelta, fromAttendee)
	// manage waitlist ugh
	// only need to manage it if the event had no spots left and spots freed up from main attendee list
	if e.SpotsLeft() == 0 && attendeeCountDelta < 0 && fromAttendee {
		waitlist, err := listWaitlist(tx, p.Id, attendeeCountDelta*-1)
		if err != nil {
			return err
		}

		err = updateWaitlist(tx, waitlist)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func get(tx *sqlx.Tx, id string) (Event, error) {
	stmt := `
        SELECT
            e.id, e.name, e.capacity, e.start, e.location, e.created_at, e.creator_id
            , u.full_name AS creator_full_name
            , COALESCE((
                SELECT SUM(attendee_count) FROM event_response
                WHERE event_id = ? AND on_waitlist = FALSE
            ), 0) AS total_attendee_count
            , e.group_id, ug.name AS group_name
            , CASE
                WHEN datetime() > datetime(start) THEN TRUE
                ELSE FALSE
            END AS is_past
        FROM event AS e
        LEFT JOIN user_group AS ug ON e.group_id = ug.id
        INNER JOIN user AS u ON e.creator_id = u.id
        WHERE e.id = ? AND e.is_deleted = FALSE 
    `
	args := []any{id, id}

	var event Event
	err := tx.Get(&event, stmt, args...)
	if err != nil {
		return Event{}, err
	}

	return event, nil
}

func listResponses(tx *sqlx.Tx, eventId string) ([]EventResponse, error) {
	stmt := `
        SELECT er.event_id, er.user_id, er.attendee_count, u.full_name AS user_full_name, er.created_at, er.on_waitlist
        FROM event_response AS er
        INNER JOIN user AS u ON er.user_id = u.id
        WHERE er.event_id = ?
        ORDER BY er.created_at
    `
	args := []any{eventId}

	rows, err := tx.Query(stmt, args...)
	if err != nil {
		return []EventResponse{}, err
	}
	defer rows.Close()
	var responses []EventResponse
	for rows.Next() {
		var i EventResponse
		if err := rows.Scan(&i.EventId, &i.UserId, &i.AttendeeCount, &i.UserFullName, &i.CreatedAt, &i.OnWaitlist); err != nil {
			return []EventResponse{}, err
		}
		responses = append(responses, i)
	}
	if err := rows.Err(); err != nil {
		return []EventResponse{}, err
	}

	return responses, nil
}

func getUserResponse(tx *sqlx.Tx, eventId string, userId string) (*EventResponse, error) {
	stmt := `
        SELECT event_id, attendee_count, on_waitlist
        FROM event_response
        WHERE event_id = ? AND user_id = ?
    `
	args := []any{eventId, userId}

	var response EventResponse
	err := tx.Get(&response, stmt, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &response, nil
}

func deleteResponse(tx *sqlx.Tx, eventId string, userId string) error {
	stmt := `
        DELETE FROM event_response
        WHERE event_id = ? AND user_id = ?
    `
	args := []any{eventId, userId}
	serviceLog("deleteResponse args %v", args)

	_, err := tx.Exec(stmt, args...)
	if err != nil {
		return err
	}

	return nil
}

type updateResponseParams struct {
	EventId       string
	UserId        string
	AttendeeCount int
	OnWaitlist    bool
}

func updateResponse(tx *sqlx.Tx, p updateResponseParams) error {
	stmt := `
        INSERT INTO event_response (event_id, user_id, created_at, updated_at, attendee_count, on_waitlist)
        VALUES (?, ?, ?, ?, ?, ?)
        ON CONFLICT (event_id, user_id) DO UPDATE SET
            updated_at = excluded.updated_at,
            attendee_count = excluded.attendee_count
    `

	now := time.Now().UTC()
	args := []any{
		p.EventId,
		p.UserId,
		now,
		now,
		p.AttendeeCount,
		p.OnWaitlist,
	}
	serviceLog("updateResponse args %v", args)

	_, err := tx.Exec(stmt, args...)
	if err != nil {
		return err
	}
	return nil
}

func listWaitlist(tx *sqlx.Tx, eventId string, limit int) ([]EventResponse, error) {
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
	err := tx.Select(&waitlist, stmt, args...)
	if err != nil {
		return []EventResponse{}, err
	}

	return waitlist, nil
}

func updateWaitlist(tx *sqlx.Tx, reqs []EventResponse) error {
	stmt := `
        UPDATE event_response
        SET on_waitlist = 0
        WHERE event_id = ? AND user_id = ?
    `

	for _, req := range reqs {
		args := []any{req.EventId, req.UserId}
		serviceLog("updateWaitlist args %v", args)

		_, err := tx.Exec(stmt, args...)
		if err != nil {
			return err
		}
	}

	return nil
}
