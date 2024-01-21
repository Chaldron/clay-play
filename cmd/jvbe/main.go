package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	fmt.Println("hello world")

	r := chi.NewRouter()

	publicFileServer := http.FileServer(http.Dir("./ui/public"))
	r.Handle("/public/*", http.StripPrefix("/public/", publicFileServer))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles(
			"./ui/views/pages/home.html",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		t.Execute(w, nil)
	})

	http.ListenAndServe(":8080", r)
}
