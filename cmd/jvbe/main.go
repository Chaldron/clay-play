package main

import (
	"flag"
	"fmt"
	"github/mattfan00/jvbe/auth"
	"github/mattfan00/jvbe/config"
	"github/mattfan00/jvbe/event"
	"github/mattfan00/jvbe/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
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

	eventStore := event.NewStore(db)
	eventService := event.NewService(eventStore, templates)

	oauthConf := &oauth2.Config{
		ClientID:     conf.FbAppId,
		ClientSecret: conf.FbSecret,
		RedirectURL:  "http://localhost:8080/auth/callback",
		Scopes:       []string{"public_profile"},
		Endpoint:     facebook.Endpoint,
	}
	authService := auth.NewService(oauthConf)

	r := chi.NewRouter()

	publicFileServer := http.FileServer(http.Dir("./ui/public"))
	r.Handle("/public/*", http.StripPrefix("/public/", publicFileServer))

	eventService.Routes(r)
	authService.Routes(r)

	http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), r)
}
