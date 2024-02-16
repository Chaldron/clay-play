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
	case "dropeverythingdangerous":
		drop()
    }

	return nil
}

func colExists(table string, col string) bool {
	_, err := db.Query(fmt.Sprintf("SELECT %s FROM %s", col, table))
	return err == nil
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
			creator_id TEXT NOT NULL,
            is_deleted BOOLEAN NOT NULL DEFAULT 0,
            group_id TEXT
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
        CREATE INDEX IF NOT EXISTS sessions_expiry_idx ON sessions(expiry);
    `)
	log.Printf("created index on table: sessions")

	db.MustExec(`
        CREATE TABLE IF NOT EXISTS event_response (
            event_id TEXT NOT NULL,
            user_id TEXT NOT NULL,
            created_at DATETIME NOT NULL,
            updated_at DATETIME NOT NULL,
            attendee_count INT NOT NULL DEFAULT 0,
            on_waitlist BOOL NOT NULL DEFAULT 0,
            PRIMARY KEY (event_id, user_id)
        )   
    `)
	log.Printf("created table: event_response")

	db.MustExec(`
        CREATE TABLE IF NOT EXISTS user_group (
            id TEXT PRIMARY KEY,
            created_at DATETIME NOT NULL,
            creator_id TEXT NOT NULL,
            is_deleted BOOL NOT NULL DEFAULT 0,
            name TEXT NOT NULL,
            invite_id TEXT NOT NULL UNIQUE
        )
    `)
	log.Printf("created table: user_group")

	db.MustExec(`
        CREATE TABLE IF NOT EXISTS user_group_member (
            group_id TEXT NOT NULL,
            user_id TEXT NOT NULL,
            created_at DATETIME NOT NULL,
            PRIMARY KEY (group_id, user_id)
        )
    `)
	log.Printf("created table: user_group_member")
}

func drop() {
	db.MustExec(`DROP TABLE IF EXISTS event`)
	log.Printf("dropped table: event")

	db.MustExec(`DROP TABLE IF EXISTS user`)
	log.Printf("dropped table: user")

	db.MustExec(`DROP TABLE IF EXISTS sessions`)
	log.Printf("dropped table: sessions")

	db.MustExec(`CREATE INDEX IF NOT EXISTS sessions_expiry_idx ON sessions(expiry);`)
	log.Printf("dropped index on table: sessions")

	db.MustExec(`DROP TABLE IF EXISTS event_response`)
	log.Printf("dropped table: event_response")

	db.MustExec(`DROP TABLE IF EXISTS user_group`)
	log.Printf("dropped table: user_group")

	db.MustExec(`DROP TABLE IF EXISTS user_group_member`)
	log.Printf("dropped table: user_group_member")
}
