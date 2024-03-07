package group

import (
	"database/sql"
	"errors"
	"time"
)

type Service interface {
	Get(string) (Group, error)
	GetDetailed(string) (GroupDetailed, error)
	List() ([]Group, error)
	CreateAndAddMember(CreateParams) error
	Update(UpdateParams) error
	Delete(string) error
	AddMemberFromInvite(string, string) (Group, error)
	RemoveMember(string, string) error
	UserCanAccess(sql.NullString, string) (bool, error)
	UserCanAccessError(sql.NullString, string) error
	RefreshInviteId(string) error
}

type Group struct {
	Id               string    `db:"id"`
	CreatedAt        time.Time `db:"created_at"`
	CreatorId        string    `db:"creator_id"`
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
