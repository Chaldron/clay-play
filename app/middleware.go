package app

import (
	"fmt"
	user "github/mattfan00/jvbe/user"
	"net/http"
)

func (a *App) getSessionUser(r *http.Request) (user.SessionUser, bool) {
	u, ok := a.session.Get(r.Context(), "user").(user.SessionUser)
	isAuth := ok && u.IsAuthenticated()
	return u, isAuth
}

func (a *App) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := a.getSessionUser(r); ok {
			next.ServeHTTP(w, r)
		} else {
			// TODO: need to handle both page requests and hx requests
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	})
}

func (a *App) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, _ := a.getSessionUser(r); u.IsAdmin {
			next.ServeHTTP(w, r)
		} else {
			// TODO: need to handle both page requests and hx requests
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}
	})
}

func (a *App) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				w.Header().Set("Connection", "close")
				http.Error(w, fmt.Errorf("%s", err).Error(), http.StatusInternalServerError)
				return
			}
		}()

		next.ServeHTTP(w, r)
	})
}
