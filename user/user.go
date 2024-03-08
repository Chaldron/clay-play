package user

import (
	"database/sql"
	"errors"
	"slices"
	"strings"
	"time"
)

type Service interface {
	Get(string) (User, error)
	HandleFromExternal(ExternalUser) (User, error)
	Create(CreateParams) (User, error)
	GetReview(string) (UserReview, error)
	UpdateReview(UpdateReviewParams) error
	ListReviews() ([]UserReview, error)
	ApproveReview(string) error
}

var (
	ErrNoUser = errors.New("no user found")
)

type UserStatus int

const (
	UserStatusActive   UserStatus = iota // has full control of application
	UserStatusInactive                   // has not been approved yet
)

type User struct {
	Id         string     `db:"id"`
	FullName   string     `db:"full_name"`
	ExternalId string     `db:"external_id"`
	CreatedAt  time.Time  `db:"created_at"`
	Status     UserStatus `db:"status"`
}

func (u *User) ToSessionUser() SessionUser {
	return SessionUser{
		Id:       u.Id,
		FullName: u.FullName,
		Status:   u.Status,
	}
}

type UserReview struct {
	UserId       string         `db:"user_id"`
	UserFullName string         `db:"user_full_name"`
	CreatedAt    time.Time      `db:"created_at"`
	ReviewedAt   sql.NullTime   `db:"reviewed_at"`
	Comment      sql.NullString `db:"comment"`
	IsApproved   string         `db:"is_approved"`
}

type ExternalUser struct {
	Id          string `json:"sub"`
	FullName    string `json:"name"`
	Picture     string `json:"picture"`
	Permissions []string
}

type SessionUser struct {
	Id          string
	Permissions []string
	FullName    string
	Status      UserStatus
}

func (u SessionUser) IsAuthenticated() bool {
	return u.Id != ""
}

func (u SessionUser) FirstName() string {
	nameParts := strings.Fields(u.FullName)
	if len(nameParts) < 1 {
		return ""
	}

	firstName := nameParts[0]
	return firstName
}

func (u SessionUser) hasPermission(p string) bool {
	return slices.Contains[[]string](u.Permissions, p)
}

func (u SessionUser) CanModifyEvent() bool {
	return u.hasPermission("modify:event")
}

func (u SessionUser) CanModifyGroup() bool {
	return u.hasPermission("modify:group")
}

func (u SessionUser) CanReviewUser() bool {
	return u.hasPermission("review:user")
}

func (u SessionUser) CanDoEverything() bool {
	return u.CanModifyGroup() &&
		u.CanModifyEvent() &&
		u.CanReviewUser()

}
