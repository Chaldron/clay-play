package main

import (
	"github/mattfan00/jvbe/event"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sqlx.Connect("sqlite3", "./jvbe.db")
	if err != nil {
		panic(err)
	}

    eventStore := event.NewStore(db)
	eventService := event.NewService(eventStore)

	r := chi.NewRouter()

	publicFileServer := http.FileServer(http.Dir("./ui/public"))
	r.Handle("/public/*", http.StripPrefix("/public/", publicFileServer))

	eventService.Routes(r)

	http.ListenAndServe(":8080", r)
}
