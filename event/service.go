package event

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/mattfan00/jvbe/db"
	"github.com/mattfan00/jvbe/logger"
)

type service struct {
	db  *db.DB
	log logger.Logger
}

func NewService(db *db.DB) *service {
	return &service{
		db:  db,
		log: logger.NewNoopLogger(),
	}
}

func (s *service) SetLogger(l logger.Logger) {
	s.log = l
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

func (s *service) GetDetailed(id string, userId int64) (EventDetailed, error) {
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

func (s *service) ListResponses(eventId string) ([]EventResponse, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return []EventResponse{}, err
	}
	defer tx.Rollback()

	el, err := listResponses(tx, eventId)
	return el, err
}

type ListFilter struct {
	UserId      int64
	Upcoming    bool
	Past        bool
	Limit       int
	Offset      int
	OrderByDesc bool
}

func (s *service) List(f ListFilter) (EventList, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return EventList{}, err
	}
	defer tx.Rollback()

	el, err := list(tx, f)
	return el, err
}

type CreateParams struct {
	Name      string
	GroupId   string
	Capacity  int
	Start     time.Time
	Location  string
	CreatorId int64
}

func (s *service) Create(p CreateParams) (string, error) {
	s.log.Printf("group Create params %+v", p)
	tx, err := s.db.Beginx()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	id, err := create(tx, p)
	if err != nil {
		return "", err
	}

	err = tx.Commit()
	if err != nil {
		return "", err
	}

	s.log.Printf("created event %s", id)
	return id, nil
}

type UpdateParams struct {
	Id       string
	Name     string
	Capacity int
	Start    time.Time
	Location string
}

func (s *service) Update(p UpdateParams) error {
	s.log.Printf("group Update params %+v", p)
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = update(tx, p)
	if err != nil {
		return err
	}

	_, err = manageWaitlist(tx, p.Id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *service) Delete(id string) error {
	s.log.Printf("group Delete id %s", id)
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

type HandleResponseParams struct {
	UserId        int64
	Id            string
	AttendeeCount int
}

func (s *service) HandleResponse(p HandleResponseParams) error {
	s.log.Printf("group HandleResponse params %+v", p)

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
		s.log.Printf("deleted response")
	} else {
		err = updateResponse(tx, updateResponseParams{
			EventId:       p.Id,
			UserId:        p.UserId,
			AttendeeCount: p.AttendeeCount,
		})
		if err != nil {
			return err
		}
	}

	_, err = manageWaitlist(tx, p.Id)
	if err != nil {
		return err
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
        INNER JOIN users AS u ON e.creator_id = u.id
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
        INNER JOIN users AS u ON er.user_id = u.id
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

func getUserResponse(tx *sqlx.Tx, eventId string, userId int64) (*EventResponse, error) {
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

func list(tx *sqlx.Tx, f ListFilter) (EventList, error) {
	where, wargs := []string{}, []any{}

	where = append(where, "is_deleted = FALSE")
	if f.Upcoming {
		where = append(where, "datetime() <= datetime(start)")
	}
	if f.Past {
		where = append(where, "datetime() > datetime(start)")
	}

	// move the logic for determining if user can access event based off group from group service over to here
	if f.UserId > -1 {
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
	err := tx.Select(&events, stmt, args...)
	if err != nil {
		return EventList{
			Events: []Event{},
		}, err
	}

	return EventList{
		Events: events,
	}, nil
}

func create(tx *sqlx.Tx, p CreateParams) (string, error) {
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

	_, err = tx.Exec(stmt, args...)
	if err != nil {
		return "", err
	}

	return newId, nil
}

func update(tx *sqlx.Tx, p UpdateParams) error {
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

	_, err := tx.Exec(stmt, args...)
	return err
}

func deleteResponse(tx *sqlx.Tx, eventId string, userId int64) error {
	stmt := `
        DELETE FROM event_response
        WHERE event_id = ? AND user_id = ?
    `
	args := []any{eventId, userId}

	_, err := tx.Exec(stmt, args...)
	if err != nil {
		return err
	}

	return nil
}

type updateResponseParams struct {
	EventId       string
	UserId        int64
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

	_, err := tx.Exec(stmt, args...)
	if err != nil {
		return err
	}
	return nil
}

// Manages the waitlist status of all attendees in an event.
// Based on the event's capacity, will convert all regular attendees to waitlist and all waitlist attendees to regular as necessary.
//
// Returns list of responses that had their waitlist status updated.
func manageWaitlist(tx *sqlx.Tx, eventId string) ([]EventResponse, error) {
	e, err := get(tx, eventId)
	if err != nil {
		return []EventResponse{}, err
	}

	stmt := `
			UPDATE event_response AS er
			SET on_waitlist = r.on_waitlist
			FROM (
				SELECT
					event_id
					,user_id
					,CASE
						WHEN SUM(attendee_count) OVER (ORDER BY created_at) <= ? THEN FALSE
						ELSE TRUE
					END AS on_waitlist
				FROM event_response
				WHERE event_id = ?
			) AS r
			WHERE er.event_id = r.event_id
				AND er.user_id = r.user_id
				AND er.on_waitlist <> r.on_waitlist
			RETURNING event_id, user_id, on_waitlist
		`
	var er []EventResponse
	err = tx.Select(&er, stmt, e.Capacity, e.Id)
	if err != nil {
		return []EventResponse{}, err
	}

	return er, nil
}
