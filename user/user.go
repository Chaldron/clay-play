package user

import (
	"errors"
	"strings"
	"time"
)

type Service interface {
	Get(int64) (User, error)
	GetAll() ([]User, error)
	HandleFromCreds(email string, password string) (User, error)
	Create(CreateParams) (User, error)
}

var (
	ErrNoUser = errors.New("no user found")
)

type User struct {
	Id        string    `db:"id"`
	FullName  string    `db:"full_name"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
	IsAdmin   bool      `db:"isadmin"`
}

func (u *User) ToSessionUser() SessionUser {
	return SessionUser{
		Id:       u.Id,
		FullName: u.FullName,
		IsAdmin:  u.IsAdmin,
	}
}

type SessionUser struct {
	Id       string
	FullName string
	IsAdmin  bool
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
