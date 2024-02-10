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
	case "mod1":
		mod1()
	}

	return nil
}

func colExists(table string, col string) bool {
	_, err := db.Query(fmt.Sprintf("SELECT %s FROM %s", col, table))
	if err == nil {
		return true
	} else {
		return false
	}
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
            external_id TEXT NOT NULL,
            created_at DATETIME NOT NULL,
            picture TEXT
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
            updated_at DATETIME NOT NULL,
            attendee_count INT NOT NULL DEFAULT 0,
            PRIMARY KEY (event_id, user_id)
        )   
    `)
	log.Printf("created table: event_response")
}

func mod1() {
	if ok := colExists("user", "is_admin"); ok {
		db.MustExec(`
            ALTER TABLE user 
            DROP COLUMN is_admin
        `)
		log.Printf("dropped is_admin from user")
	}

	if ok := colExists("event_response", "going"); ok {
		db.MustExec(`
            ALTER TABLE event_response
            DROP COLUMN going
        `)
		log.Printf("dropped going from event_response")
	}

	if ok := colExists("event_response", "attendee_count"); !ok {
		db.MustExec(`
            ALTER TABLE event_response
            ADD COLUMN attendee_count INT NOT NULL DEFAULT 0
        `)
		log.Printf("added attendee_count to event_response")
	}
}
