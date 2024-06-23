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

func (a *App) handleLoginCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("login callback: %s", r.URL.String())

		//state := r.URL.Query().Get("state")
		//expectedState := a.session.PopString(r.Context(), "state")
		//if state != expectedState {
		//	err := fmt.Errorf("invalid oauth state, expected '%s', got '%s'", expectedState, state)
		//	a.renderErrorPage(w, err, http.StatusInternalServerError)
		//	return
		//}

		u, err := a.userService.HandleFromCreds("123@abc.com", "pass")
		if err != nil {
			a.renderErrorPage(w, err, http.StatusInternalServerError)
			return
		}
		sessionUser := u.ToSessionUser()
		sessionUser.Permissions = []string{"modify:event", "modify:group", "review:user"}

		log.Printf("sessionUser:%+v", sessionUser)

		if err := a.renewSessionUser(r, &sessionUser); err != nil {
			a.renderErrorPage(w, err, http.StatusInternalServerError)
			return
		}

		redirect := a.session.PopString(r.Context(), "redirect")
		if redirect != "" {
			http.Redirect(w, r, redirect, http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/home", http.StatusSeeOther)
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
