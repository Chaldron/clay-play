package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

var DefaultAdminPassword string

func init() {
	goose.AddMigration(Up, Down)
}

func Up(tx *sql.Tx) error {
	_, err := tx.Exec(
		`INSERT INTO users (id, full_name, email, password, isadmin, created_at)
		VALUES (0, "Admin", "admin@example.com", "` +
			DefaultAdminPassword +
			`", true, CURRENT_TIMESTAMP)`)

	if err != nil {
		return err
	}
	return nil
}

func Down(tx *sql.Tx) error {
	_, err := tx.Exec("UPDATE users SET username='root' WHERE username='admin';")
	if err != nil {
		return err
	}
	return nil
}
