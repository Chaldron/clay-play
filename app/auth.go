package app

import (
	"log"
	"net/http"
	"strings"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

func (a *App) renderLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, err := gonanoid.New()
		if err != nil {
			a.renderErrorPage(w, err, http.StatusInternalServerError)
		}

		a.session.Put(r.Context(), "state", state)

		http.Redirect(w, r, "/auth/callback", http.StatusSeeOther)
	}
}

func (app *App) handleLoginCallback() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		log.Printf("login callback: %s", request.URL.String())

		email := strings.ToLower(request.URL.Query().Get("email"))
		password := request.URL.Query().Get("password")

		user, err := app.userService.HandleFromCreds(email, password)
		if err != nil {
			app.renderErrorPage(response, err, http.StatusInternalServerError)
			return
		}

		sessionUser := user.ToSessionUser()

		log.Printf("sessionUser:%+v", sessionUser)

		if err := app.renewSessionUser(request, &sessionUser); err != nil {
			app.renderErrorPage(response, err, http.StatusInternalServerError)
			return
		}

		redirect := app.session.PopString(request.Context(), "redirect")
		if redirect != "" {
			http.Redirect(response, request, redirect, http.StatusSeeOther)
		} else {
			http.Redirect(response, request, "/home", http.StatusSeeOther)
		}
	}
}

func (app *App) handleLogout() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		err := app.session.Destroy(request.Context())
		if err != nil {
			app.renderErrorPage(response, err, http.StatusInternalServerError)
			return
		}
		http.Redirect(response, request, "/", http.StatusSeeOther)
	}
}
