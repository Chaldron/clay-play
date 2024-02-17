package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	appPkg "github/mattfan00/jvbe/app"
	"github/mattfan00/jvbe/auth"
	"github/mattfan00/jvbe/config"
	"github/mattfan00/jvbe/event"
	"github/mattfan00/jvbe/group"
	"github/mattfan00/jvbe/template"
	"github/mattfan00/jvbe/user"
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

	templates, err := template.Generate()
	if err != nil {
		return err
	}

	gob.Register(user.SessionUser{}) // needed for scs library
	session := scs.New()
	session.Lifetime = 30 * 24 * time.Hour // 30 days
	session.Store = sqlite3store.New(db.DB)

	groupStore := group.NewStore(db)
	groupService := group.NewService(groupStore, conf)

	eventStore := event.NewStore(db)
	eventService := event.NewService(eventStore, groupService)

	userStore := user.NewStore(db)
	userService := user.NewService(userStore)

	authService, err := auth.NewService(conf, userService)
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
