package event

import (
	"database/sql"
	"time"
)

type Service interface {
	Get(string) (Event, error)
	GetDetailed(string, int64) (EventDetailed, error)
	ListResponses(string) ([]EventResponse, error)
	List(ListFilter) (EventList, error)
	Create(CreateParams) (string, error)
	Update(UpdateParams) error
	Delete(string) error
	HandleResponse(HandleResponseParams) error
}

type Event struct {
	Id                 string         `db:"id"`
	Name               string         `db:"name"`
	GroupId            sql.NullString `db:"group_id"`
	GroupName          sql.NullString `db:"group_name"`
	Capacity           int            `db:"capacity"`
	Start              time.Time      `db:"start"`
	Location           string         `db:"location"`
	CreatedAt          time.Time      `db:"created_at"`
	CreatorId          string         `db:"creator_id"`
	CreatorFullName    string         `db:"creator_full_name"`
	TotalAttendeeCount int            `db:"total_attendee_count"`
	IsPast             bool           `db:"is_past"`
}

func (e Event) SpotsLeft() int {
	return e.Capacity - e.TotalAttendeeCount
}

type EventResponse struct {
	EventId       string    `db:"event_id"`
	UserId        int64     `db:"user_id"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
	AttendeeCount int       `db:"attendee_count"`
	OnWaitlist    bool      `db:"on_waitlist"`
	UserFullName  string    `db:"user_full_name"`
}

func (e EventResponse) PlusOnes() int {
	return e.AttendeeCount - 1
}

type EventDetailed struct {
	Event
	UserResponse *EventResponse
	Responses    []EventResponse
}

// containing this in a struct in case need to include more fields for pagination
type EventList struct {
	Events []Event
}

var MaxAttendeeCount = 2
