package user

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/Chaldron/clay-play/db"
	"github.com/Chaldron/clay-play/logger"
	"github.com/jmoiron/sqlx"
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

func (s *service) Get(id int64) (User, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()

	u, err := get(tx, id)
	return u, err
}

func (s *service) GetAll() ([]User, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	u, err := getAll(tx)
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

func (s *service) Delete(id int64) error {
	s.log.Printf("user Delete id %s", id)
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = delete(tx, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

type CreateParams struct {
	FullName string
	Email    string
	Password string
	IsAdmin  bool
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

type UpdateParams struct {
	Id       int64
	FullName string
	Email    string
	Password string
	IsAdmin  bool
}

func (s *service) Update(p UpdateParams) (User, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()

	u, err := update(tx, p)
	if err != nil {
		return User{}, err
	}

	err = tx.Commit()
	if err != nil {
		return User{}, err
	}

	return u, err
}

func get(tx *sqlx.Tx, id int64) (User, error) {
	stmt := `
        SELECT id, full_name, email, created_at, isadmin FROM users
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

func getAll(tx *sqlx.Tx) ([]User, error) {
	stmt := `
        SELECT id, full_name, email, created_at, isadmin FROM users
    `

	var users []User
	err := tx.Select(&users, stmt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoUser
	} else if err != nil {
		return nil, err
	}

	return users, nil
}

func getByExternal(tx *sqlx.Tx, email string, password string) (User, error) {
	stmt := `
        SELECT 
            id, full_name, created_at, email, isadmin
        FROM users
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
	stmt := `
        INSERT INTO users (full_name, email, password, created_at, isadmin)
        VALUES (?, ?, ?, ?, ?)
    `
	args := []any{
		p.FullName,
		strings.ToLower(p.Email),
		p.Password,
		time.Now().UTC(),
		p.IsAdmin,
	}

	u, err := tx.Exec(stmt, args...)
	if err != nil {
		return User{}, err
	}

	newId, _ := u.LastInsertId()
	newUser, err := get(tx, newId)
	if err != nil {
		return User{}, err
	}

	return newUser, nil
}

func update(tx *sqlx.Tx, p UpdateParams) (User, error) {
	stmt := `
		UPDATE users
		SET 
			full_name = ?, 
			email = ?, 
			password = CASE WHEN ? = '' THEN password ELSE ? END, 
			isadmin = ?
		WHERE id = ?
		`
	args := []any{
		p.FullName,
		p.Email,
		p.Password,
		p.Password,
		p.IsAdmin,
		p.Id,
	}

	_, err := tx.Exec(stmt, args...)
	if err != nil {
		return User{}, err
	}

	newUser, err := get(tx, p.Id)
	if err != nil {
		return User{}, err
	}

	return newUser, nil
}

func delete(tx *sqlx.Tx, id int64) error {
	stmt := `
        DELETE FROM users
        WHERE id = ?
    `
	args := []any{id}

	_, err := tx.Exec(stmt, args...)
	return err
}
