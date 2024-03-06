package event

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Service interface {
	Get(string) (Event, error)
	GetDetailed(string, string) (EventDetailed, error)
	ListCurrent() ([]Event, error)
	Create(CreateParams) error
	Update(UpdateParams) error
	Delete(string) error
	HandleResponse(HandleResponseParams) error
}

func serviceLog(format string, s ...any) {
	log.Printf("event/service.go: %s", fmt.Sprintf(format, s...))
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
}

func (e Event) SpotsLeft() int {
	return e.Capacity - e.TotalAttendeeCount
}

type EventResponse struct {
	EventId       string    `db:"event_id"`
	UserId        string    `db:"user_id"`
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

var MaxAttendeeCount = 2
