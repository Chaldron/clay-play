package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"net/http"
	"time"

	appPkg "github.com/mattfan00/jvbe/app"
	"github.com/mattfan00/jvbe/auditlog"
	"github.com/mattfan00/jvbe/auth"
	"github.com/mattfan00/jvbe/config"
	"github.com/mattfan00/jvbe/db"
	"github.com/mattfan00/jvbe/event"
	"github.com/mattfan00/jvbe/group"
	"github.com/mattfan00/jvbe/logger"
	"github.com/mattfan00/jvbe/template"
	"github.com/mattfan00/jvbe/user"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
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
	log := logger.NewStdLogger()

	conf, err := config.ReadFile(p.configPath)
	if err != nil {
		return err
	}

	db, err := db.Connect(conf.DbConn, log)
	if err != nil {
		return err
	}

	templates, err := template.Generate()
	if err != nil {
		return err
	}

	gob.Register(user.SessionUser{}) // needed for scs library
	session := scs.New()
	session.Lifetime = 30 * 24 * time.Hour // 30 days
	session.Store = sqlite3store.New(db.DB.DB)

	groupService := group.NewService(db)
	groupService.SetLogger(log)

	eventService := event.NewService(db)
	eventService.SetLogger(log)

	userService := user.NewService(db)
	eventService.SetLogger(log)

	auditlogService := auditlog.NewService(db)

	authService, err := auth.NewService()
	if err != nil {
		return err
	}
	authService.SetLogger(log)

	app := appPkg.New(
		eventService,
		userService,
		authService,
		groupService,
		auditlogService,

		conf,
		session,
		templates,
		log,
	)

	log.Printf("listening on port %d", conf.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), app.Routes())

	return nil
}
