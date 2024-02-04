package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	appPkg "github/mattfan00/jvbe/app"
	"github/mattfan00/jvbe/auth"
	"github/mattfan00/jvbe/config"
	"github/mattfan00/jvbe/event"
	"github/mattfan00/jvbe/facebook"
	"github/mattfan00/jvbe/template"
	"github/mattfan00/jvbe/user"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"
	oauthFacebook "golang.org/x/oauth2/facebook"
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
	session.Lifetime = 7 * 24 * time.Hour // 1 week
	session.Store = sqlite3store.New(db.DB)

	eventStore := event.NewStore(db)
	eventService := event.NewService(eventStore, templates)

	userStore := user.NewStore(db)
	userService := user.NewService(userStore)

	oauthConf := &oauth2.Config{
		ClientID:     conf.FbAppId,
		ClientSecret: conf.FbSecret,
		RedirectURL:  "http://localhost:8080/auth/callback",
		Scopes:       []string{"public_profile"},
		Endpoint:     oauthFacebook.Endpoint,
	}
	facebookService := facebook.NewService(oauthConf)

	authService := auth.NewService(userService, facebookService)

	app := appPkg.New(
		eventService,
		userService,
		authService,

		session,
		templates,
	)

	log.Printf("listening on port %d", conf.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), app.Routes())

	return nil
}
