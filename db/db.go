package db

import (
	"embed"
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

type DB struct {
	*sqlx.DB
}

//go:embed migrations/*.sql
var migrationsFs embed.FS

func Connect(dsn string) (*DB, error) {
	db, err := sqlx.Connect("sqlite3", dsn)
	if err != nil {
		return &DB{}, err
	}
	log.Print("connected to db: ", dsn)

	goose.SetBaseFS(migrationsFs)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, err
	}

	newDB := &DB{
		DB: db,
	}

	log.Print("beginning migration")
	if err := newDB.MigrationUp(); err != nil {
		return &DB{}, err
	}

	return newDB, nil
}

func TestingConnect(t testing.TB) *DB {
	t.Helper()
	db, err := Connect(":memory:")
	if err != nil {
		panic(err)
	}
	return db
}

func (db *DB) MigrationCreate(name string) error {
	if name == "" {
		return errors.New("provide a name for the migration")
	}
	return goose.Create(db.DB.DB, "./db/migrations", name, "sql")
}

func (db *DB) MigrationUp() error {
	return goose.Up(db.DB.DB, "migrations")
}

func (db *DB) MigrationDownTo(version int64) error {
	return goose.DownTo(db.DB.DB, "migrations", version)
}

// FormatLimitOffset returns a SQL string for a given limit & offset.
// Clauses are only added if limit and/or offset are greater than zero.
func FormatLimitOffset(limit, offset int) string {
	if limit > 0 && offset > 0 {
		return fmt.Sprintf(`LIMIT %d OFFSET %d`, limit, offset)
	} else if limit > 0 {
		return fmt.Sprintf(`LIMIT %d`, limit)
	} else if offset > 0 {
		return fmt.Sprintf(`OFFSET %d`, offset)
	}
	return ""
}
