package main

import (
	"flag"
	"fmt"
	"github/mattfan00/jvbe/config"
	"log"

	"github.com/jmoiron/sqlx"
)

type dbProgram struct {
	fs         *flag.FlagSet
	args       []string
	configPath string
}

func newDbProgram(args []string) *dbProgram {
	fs := flag.NewFlagSet("db", flag.ExitOnError)
	p := &dbProgram{
		fs:   fs,
		args: args,
	}

	fs.StringVar(&p.configPath, "c", "./config.yaml", "path to config file")

	return p
}

func (p *dbProgram) parse() error {
	return p.fs.Parse(p.args)
}

func (p *dbProgram) name() string {
	return p.fs.Name()
}

var db *sqlx.DB

func (p *dbProgram) run() error {
	if len(p.fs.Args()) != 1 {
		return fmt.Errorf("incorrect num of args")
	}

	conf, err := config.ReadFile(p.configPath)
	if err != nil {
		return err
	}

	action := p.fs.Arg(0)

	db = sqlx.MustConnect("sqlite3", conf.DbConn)
	log.Printf("connected to DB: %s\n", conf.DbConn)

	switch action {
	case "create":
		create()
	}

	return nil
}

func create() {
	db.MustExec(`
        CREATE TABLE IF NOT EXISTS event (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            capacity INTEGER NOT NULL,
            start DATETIME NOT NULL,
            location TEXT NOT NULL,
            created_at DATETIME NOT NULL,
            is_deleted BOOLEAN NOT NULL DEFAULT 0
        )
    `)
	log.Printf("created table: event")

	db.MustExec(`
        CREATE TABLE IF NOT EXISTS user (
            id TEXT PRIMARY KEY,
            full_name TEXT NOT NULL,
            external_id TEXT,
            is_admin BOOLEAN NOT NULL,
            created_at DATETIME NOT NULL
        )

    `)
	log.Printf("created table: user")

	db.MustExec(`
        CREATE TABLE IF NOT EXISTS sessions (
            token TEXT PRIMARY KEY,
            data BLOB NOT NULL,
            expiry REAL NOT NULL
        )

    `)
	log.Printf("created table: sessions")

	db.MustExec(`
        CREATE INDEX sessions_expiry_idx ON sessions(expiry);
    `)
	log.Printf("created index on table: sessions")

	db.MustExec(`
        CREATE TABLE IF NOT EXISTS event_response (
            event_id TEXT NOT NULL,
            user_id TEXT NOT NULL,
            going BOOLEAN NOT NULL,
            updated_at DATETIME NOT NULL,
            PRIMARY KEY (event_id, user_id)
        )   
    `)
	log.Printf("created table: event_response")
}
