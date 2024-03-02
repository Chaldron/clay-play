package main

import (
	"database/sql"
	"errors"
	"flag"
	"github.com/mattfan00/jvbe/config"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

type migrationProgram struct {
	fs         *flag.FlagSet
	args       []string
	configPath string
	migration  *migration
}

func newMigrationProgram(args []string) *migrationProgram {
	fs := flag.NewFlagSet("migration", flag.ExitOnError)
	p := &migrationProgram{
		fs:   fs,
		args: args,
	}

	fs.StringVar(&p.configPath, "c", "./config.yaml", "path to config file")

	return p
}

func (p *migrationProgram) parse() error {
	return p.fs.Parse(p.args)
}

func (p *migrationProgram) name() string {
	return p.fs.Name()
}

func (p *migrationProgram) run() error {
	action := p.fs.Arg(0)
	if action == "" {
		return errors.New("provide an action")
	}

	conf, err := config.ReadFile(p.configPath)
	if err != nil {
		return err
	}

    db := sqlx.MustConnect("sqlite3", conf.DbConn)
	log.Printf("connected to DB: %s\n", conf.DbConn)

	p.migration, err = newMigration(db.DB)
	if err != nil {
		return err
	}

	switch action {
	case "create":
		return p.create(p.fs.Arg(1))
	}

	return nil
}

func (p *migrationProgram) create(name string) error {
	return p.migration.Create(name)
}

type migration struct {
	db  *sql.DB
	dir string
}

func newMigration(db *sql.DB) (*migration, error) {
	dir := "./migrations"
	goose.SetBaseFS(os.DirFS(dir))
	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, err
	}

	return &migration{
		db:  db,
		dir: dir,
	}, nil
}

func (m *migration) Create(name string) error {
	if name == "" {
		return errors.New("provide a name for the migration")
	}
	return goose.Create(m.db, m.dir, name, "sql")
}

func (m *migration) Up() error {
	return goose.Up(m.db, ".")
}
