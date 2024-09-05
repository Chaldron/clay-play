package main

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"os"
	"time"

	appPkg "github.com/Chaldron/clay-play/app"
	"github.com/Chaldron/clay-play/auditlog"
	"github.com/Chaldron/clay-play/config"
	"github.com/Chaldron/clay-play/db"
	"github.com/Chaldron/clay-play/event"
	"github.com/Chaldron/clay-play/group"
	"github.com/Chaldron/clay-play/logger"
	"github.com/Chaldron/clay-play/template"
	"github.com/Chaldron/clay-play/user"
	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	_ "github.com/mattn/go-sqlite3"
)

func run() error {
	log := logger.NewStdLogger()

	conf, err := config.LoadFromCommandLineArgs(os.Args[1:])
	if err != nil {
		return err
	}

	db, err := db.Connect(conf.DbConn, conf.DefaultAdminPassword, log)
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

	app := appPkg.New(
		eventService,
		userService,
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

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
