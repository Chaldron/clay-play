package group

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Chaldron/clay-play/event"
)

type Service interface {
	Get(string) (Group, error)
	GetDetailed(string) (GroupDetailed, error)
	List() ([]Group, error)
	CreateAndAddMember(CreateParams) (string, error)
	Update(UpdateParams) error
	Delete(string) error
	AddMemberFromInvite(string, int64) (Group, error)
	RemoveMember(string, int64) error
	UserCanAccess(sql.NullString, int64) (bool, error)
	UserCanAccessError(sql.NullString, int64) error
	FilterEventsUserCanAccess([]event.Event, int64) ([]event.Event, error)
	RefreshInviteId(string) error
}

type Group struct {
	Id               string    `db:"id"`
	CreatedAt        time.Time `db:"created_at"`
	CreatorId        int64     `db:"creator_id"`
	CreatorFullName  string    `db:"creator_full_name"`
	IsDeleted        bool      `db:"is_deleted"`
	Name             string    `db:"name"`
	InviteId         string    `db:"invite_id"`
	TotalMemberCount int       `db:"total_member_count"`
}

type GroupMember struct {
	GroupId      string    `db:"group_id"`
	UserId       string    `db:"user_id"`
	UserFullName string    `db:"user_full_name"`
	CreatedAt    time.Time `db:"created_at"`
}

type GroupDetailed struct {
	Group
	Members []GroupMember
}

var (
	ErrNoAccess = errors.New("you do not have access")
)
