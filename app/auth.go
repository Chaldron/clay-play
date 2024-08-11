package app

import (
	"log"
	"net/http"

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

		email := request.URL.Query().Get("email")
		password := request.URL.Query().Get("password")

		user, err := app.userService.HandleFromCreds(email, password) // "123@abc.com", "pass")
		if err != nil {
			app.renderErrorPage(response, err, http.StatusInternalServerError)
			return
		}

		sessionUser := user.ToSessionUser()
		sessionUser.Permissions = []string{"modify:event", "modify:group", "review:user"}

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

func (a *App) handleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := a.session.Destroy(r.Context())
		if err != nil {
			a.renderErrorPage(w, err, http.StatusInternalServerError)
			return
		}

		//redirect := fmt.Sprintf("https://%s/logout?redirect=%s", a.conf.Oauth.Domain, a.conf.OauthLogoutRedirectUrl())
		//http.Redirect(w, r, redirect, http.StatusSeeOther)
	}
}
