package user

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/matoous/go-nanoid/v2"
)

type Store interface {
	GetByExternal(string) (User, error)
	Get(string) (User, error)
	CreateFromExternal(ExternalUser) (User, error)
}

type store struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) *store {
	return &store{
		db: db,
	}
}

func storeLog(format string, s ...any) {
	log.Printf("user/store.go: %s", fmt.Sprintf(format, s...))
}

var (
	ErrNoUser = errors.New("no user found")
)

type User struct {
	Id         string    `db:"id"`
	FullName   string    `db:"full_name"`
	ExternalId string    `db:"external_id"`
	CreatedAt  time.Time `db:"created_at"`
}

func (u *User) ToSessionUser() SessionUser {
	return SessionUser{
		Id:       u.Id,
		FullName: u.FullName,
	}
}

type ExternalUser struct {
	Id       string `json:"sub"`
	FullName string `json:"name"`
	Picture  string `json:"picture"`
}

type SessionUser struct {
	Id          string
	Permissions []string
	FullName    string
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

func (s *store) GetByExternal(externalId string) (User, error) {
	stmt := `
        SELECT id, full_name, external_id, created_at FROM user
        WHERE external_id = ?
    `
	args := []any{externalId}

	var user User
	err := s.db.Get(&user, stmt, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNoUser
	} else if err != nil {
		return User{}, err
	}

	return user, nil
}

func (s *store) Get(id string) (User, error) {
	stmt := `
        SELECT id, full_name, external_id, created_at FROM user
        WHERE id = ?
    `
	args := []any{id}

	var user User
	err := s.db.Get(&user, stmt, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNoUser
	} else if err != nil {
		return User{}, err
	}

	return user, nil
}

func (s *store) CreateFromExternal(externalUser ExternalUser) (User, error) {
	newId, err := gonanoid.New()
	if err != nil {
		return User{}, err
	}

	stmt := `
        INSERT INTO user (id, full_name, external_id, created_at)
        VALUES (?, ?, ?, ?)
    `
	args := []any{
		newId,
		externalUser.FullName,
		externalUser.Id,
		time.Now().UTC(),
	}
	storeLog("CreateFromExternal args %v", args)

	_, err = s.db.Exec(stmt, args...)
	if err != nil {
		return User{}, err
	}

	newUser, err := s.Get(newId)
	if err != nil {
		return User{}, err
	}

	return newUser, nil
}
