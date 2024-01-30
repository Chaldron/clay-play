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
	"net/http"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"
	oauthFacebook "golang.org/x/oauth2/facebook"
)

func main() {
	configPath := flag.String("c", "./config.yaml", "path to config file")
	conf, err := config.ReadFile(*configPath)
	if err != nil {
		panic(err)
	}

	db, err := sqlx.Connect("sqlite3", conf.DbConn)
	if err != nil {
		panic(err)
	}

	templates, err := template.Generate()
	if err != nil {
		panic(err)
	}

	gob.Register(user.SessionUser{}) // needed for scs library
	session := scs.New()
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

	http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), app.Routes())
}
