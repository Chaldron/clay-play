package main

import (
	"errors"
	"flag"
	"strconv"

	"github.com/Chaldron/clay-play/config"
	"github.com/Chaldron/clay-play/db"
	"github.com/Chaldron/clay-play/logger"
)

type migrationProgram struct {
	fs         *flag.FlagSet
	args       []string
	configPath string
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

	log := logger.NewStdLogger()
	db, err := db.Connect(conf.DbConn, conf.DefaultAdminPassword, log)
	if err != nil {
		return err
	}
	log.Printf("connected to DB: %s", conf.DbConn)

	switch action {
	case "create":
		return db.MigrationCreate(p.fs.Arg(1))
	case "down-to":
		v, err := strconv.ParseInt(p.fs.Arg(1), 10, 64)
		if err != nil {
			return err
		}
		return db.MigrationDownTo(v)
	}

	return nil
}
