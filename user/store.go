package user

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/matoous/go-nanoid/v2"
)

type User struct {
	Id         string    `db:"id"`
	FullName   string    `db:"full_name"`
	ExternalId string    `db:"external_id"`
	CreatedAt  time.Time `db:"created_at"`
}

type ExternalUser struct {
	Id       string
	FullName string
}

type SessionUser struct {
	Id string
}

type Store struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) GetIdByExternalId(externalId string) (string, error) {
	stmt := `
        SELECT id FROM user
        WHERE external_id = ?
    `
	args := []any{externalId}

	var id string
	err := s.db.QueryRow(stmt, args...).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return id, err
}

func (s *Store) CreateFromExternal(externalUser ExternalUser) (string, error) {
	newId, err := gonanoid.New()
	if err != nil {
		return "", err
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

	_, err = s.db.Exec(stmt, args...)
	if err != nil {
		return "", err
	}

	return newId, nil
}
