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
	GetReview(string) (UserReview, error)
	UpdateReview(UpdateReviewParams) error
	ListReviews() ([]UserReview, error)
	ApproveReview(string) error
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
	Id       string `json:"sub"`
	FullName string `json:"name"`
	Picture  string `json:"picture"`
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

func (s *store) GetByExternal(externalId string) (User, error) {
	stmt := `
        SELECT 
            id, full_name, external_id, created_at, status
        FROM user
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
        SELECT id, full_name, external_id, created_at, status FROM user
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

func (s *store) GetReview(userId string) (UserReview, error) {
	stmt := `
        SELECT user_id, comment FROM user_review
        WHERE user_id = ?
    `
	args := []any{userId}

	var userReview UserReview
	err := s.db.Get(&userReview, stmt, args...)
	return userReview, err
}

type UpdateReviewParams struct {
	UserId    string
	CreatedAt time.Time
	Comment   sql.NullString
}

func (s *store) UpdateReview(p UpdateReviewParams) error {
	stmt := `
        INSERT INTO user_review (user_id, created_at, comment)
        VALUES (?, ?, ?)
        ON CONFLICT (user_id) DO UPDATE SET
            comment = excluded.comment
    `
	args := []any{
		p.UserId,
		time.Now().UTC(),
		p.Comment,
	}
	storeLog("CreateReview args %v", args)

	_, err := s.db.Exec(stmt, args...)
	return err
}

// TODO: use transactions
func (s *store) CreateFromExternal(externalUser ExternalUser) (User, error) {
	newId, err := gonanoid.New()
	if err != nil {
		return User{}, err
	}

	stmt := `
        INSERT INTO user (id, full_name, external_id, created_at, status)
        VALUES (?, ?, ?, ?, ?)
    `
	args := []any{
		newId,
		externalUser.FullName,
		externalUser.Id,
		time.Now().UTC(),
		UserStatusInactive,
	}
	storeLog("CreateFromExternal args %v", args)

	_, err = s.db.Exec(stmt, args...)
	if err != nil {
		return User{}, err
	}

	err = s.UpdateReview(UpdateReviewParams{
		UserId:  newId,
		Comment: sql.NullString{},
	})
	if err != nil {
		return User{}, err
	}

	newUser, err := s.Get(newId)
	if err != nil {
		return User{}, err
	}

	return newUser, nil
}

func (s *store) ListReviews() ([]UserReview, error) {
	stmt := `
        SELECT 
            ur.user_id, ur.created_at, ur.comment 
            , u.full_name AS user_full_name
        FROM user_review ur
        INNER JOIN user u ON ur.user_id = u.id
        WHERE is_approved = 0
        ORDER BY ur.created_at
    `

	var u []UserReview
	err := s.db.Select(&u, stmt)
	return u, err
}

func (s *store) ApproveReview(userId string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt := `
        UPDATE user_review
        SET is_approved = 1, reviewed_at = ?
        WHERE user_id = ?
    `
	args := []any{time.Now().UTC(), userId}

	_, err = tx.Exec(stmt, args...)
	if err != nil {
		return err
	}
	storeLog("ApproveReview: approved %s", userId)

	stmt = `
        UPDATE user
        SET status = ?
        WHERE id = ?
    `
	args = []any{UserStatusActive, userId}

	_, err = tx.Exec(stmt, args...)
	if err != nil {
		return err
	}
	storeLog("ApproveReview: set %s to active status", userId)

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
