package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	appPkg "github.com/mattfan00/jvbe/app"
	"github.com/mattfan00/jvbe/auth"
	"github.com/mattfan00/jvbe/config"
	"github.com/mattfan00/jvbe/event"
	"github.com/mattfan00/jvbe/group"
	"github.com/mattfan00/jvbe/template"
	"github.com/mattfan00/jvbe/user"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type appProgram struct {
	fs         *flag.FlagSet
	args       []string
	configPath string
}

func newAppProgram(args []string) *appProgram {
	fs := flag.NewFlagSet("app", flag.ExitOnError)
	p := &appProgram{
		fs:   fs,
		args: args,
	}

	fs.StringVar(&p.configPath, "c", "./config.yaml", "path to config file")

	return p
}

func (p *appProgram) parse() error {
	return p.fs.Parse(p.args)
}

func (p *appProgram) name() string {
	return p.fs.Name()
}

func (p *appProgram) run() error {
	conf, err := config.ReadFile(p.configPath)
	if err != nil {
		return err
	}

	db, err := sqlx.Connect("sqlite3", conf.DbConn)
	if err != nil {
		return err
	}
	log.Print("connected to db: ", conf.DbConn)

	log.Print("beginning migration")
	migration, err := newMigration(db.DB)
	if err != nil {
		return err
	}
	if err := migration.Up(); err != nil {
		return err
	}

	templates, err := template.Generate()
	if err != nil {
		return err
	}

	gob.Register(user.SessionUser{}) // needed for scs library
	session := scs.New()
	session.Lifetime = 30 * 24 * time.Hour // 30 days
	session.Store = sqlite3store.New(db.DB)

	groupService := group.NewService(db)
	eventService := event.NewService(db)
	userService := user.NewService(db)

	authService, err := auth.NewService(conf)
	if err != nil {
		return err
	}

	app := appPkg.New(
		eventService,
		userService,
		authService,
		groupService,

		conf,
		session,
		templates,
	)

	log.Printf("listening on port %d", conf.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), app.Routes())

	return nil
}
