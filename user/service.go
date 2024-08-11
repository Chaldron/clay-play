package user

import (
	"database/sql"
	"errors"
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

func (s *service) Get(id string) (User, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()

	u, err := get(tx, id)
	return u, err
}

func (s *service) HandleFromCreds(email string, password string) (User, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()

	user, err := getByExternal(tx, email, password)
	if err != nil {
		return User{}, err
	}

	err = tx.Commit()
	if err != nil {
		return User{}, err
	}

	return user, nil
}

type CreateParams struct {
	ExternalId string
	FullName   string
	Picture    string
	Email      string
	Password   string
}

func (s *service) Create(p CreateParams) (User, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()

	u, err := create(tx, p)
	if err != nil {
		return User{}, err
	}

	err = tx.Commit()
	if err != nil {
		return User{}, err
	}

	return u, err
}

func get(tx *sqlx.Tx, id string) (User, error) {
	stmt := `
        SELECT id, full_name, email, created_at FROM user
        WHERE id = ?
    `
	args := []any{id}

	var user User
	err := tx.Get(&user, stmt, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNoUser
	} else if err != nil {
		return User{}, err
	}

	return user, nil
}

func getByExternal(tx *sqlx.Tx, email string, password string) (User, error) {
	stmt := `
        SELECT 
            id, full_name, created_at, email, isadmin
        FROM user
        WHERE email = ? AND password = ?
    `
	args := []any{email, password}

	var user User
	err := tx.Get(&user, stmt, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNoUser
	} else if err != nil {
		return User{}, err
	}

	return user, nil
}

func create(tx *sqlx.Tx, p CreateParams) (User, error) {
	newId, err := gonanoid.New()
	if err != nil {
		return User{}, err
	}

	stmt := `
        INSERT INTO user (id, full_name, email, password, created_at)
        VALUES (?, ?, ?, ?, ?)
    `
	args := []any{
		newId,
		p.FullName,
		p.Email,
		p.Password,
		time.Now().UTC(),
	}

	_, err = tx.Exec(stmt, args...)
	if err != nil {
		return User{}, err
	}

	newUser, err := get(tx, newId)
	if err != nil {
		return User{}, err
	}

	return newUser, nil
}
