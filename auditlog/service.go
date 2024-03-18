package auditlog

import (
	"github.com/jmoiron/sqlx"
	"github.com/mattfan00/jvbe/db"
)

type service struct {
	db *db.DB
}

func NewService(db *db.DB) *service {
	return &service{
		db: db,
	}
}

func (s *service) Create(userId string, description string) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = create(tx, userId, description)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *service) List() ([]AuditLog, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return []AuditLog{}, err
	}
	defer tx.Rollback()

	al, err := list(tx)
	return al, err
}

var al = []AuditLog{}

func create(tx *sqlx.Tx, userId string, description string) error {
	stmt := `
        INSERT INTO audit_log (user_id, recorded_at, description)
        VALUES (?, ?, ?)
    `
	args := []any{
		userId,
		db.Now(),
		description,
	}

	_, err := tx.Exec(stmt, args...)
	return err
}

func list(tx *sqlx.Tx) ([]AuditLog, error) {
	stmt := `
        SELECT 
            user_id
            ,u.full_name AS user_full_name
            ,recorded_at
            ,description 
        FROM audit_log al
        INNER JOIN user u ON al.user_id = u.id
        ORDER BY recorded_at DESC
    `

	var al []AuditLog
	err := tx.Select(&al, stmt)
	return al, err
}
