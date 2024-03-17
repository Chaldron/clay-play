package app

import (
	"errors"
	"fmt"
	user "github.com/mattfan00/jvbe/user"
	"net/http"
)

func (a *App) sessionUser(r *http.Request) (user.SessionUser, bool) {
	u, ok := a.session.Get(r.Context(), "user").(user.SessionUser)
	o := ok && u.IsAuthenticated()
	return u, o
}

// TODO: clear sessions if auth fails
func (a *App) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := a.sessionUser(r)

		if ok {
			if u.Status != user.UserStatusActive {
				http.Redirect(w, r, "/review/request", http.StatusSeeOther)
				return
			} else {
				next.ServeHTTP(w, r)
				return
			}
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	})
}

func (a *App) canModifyGroup(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, _ := a.sessionUser(r); u.CanModifyGroup() {
			next.ServeHTTP(w, r)
		} else {
			status := http.StatusUnauthorized
			a.renderErrorPage(w, errors.New(http.StatusText(status)), status)
			return
		}
	})
}

func (a *App) canModifyEvent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, _ := a.sessionUser(r); u.CanModifyEvent() {
			next.ServeHTTP(w, r)
		} else {
			status := http.StatusUnauthorized
			a.renderErrorPage(w, errors.New(http.StatusText(status)), status)
			return
		}
	})
}

func (a *App) canReviewUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, _ := a.sessionUser(r); u.CanReviewUser() {
			next.ServeHTTP(w, r)
		} else {
			status := http.StatusUnauthorized
			a.renderErrorPage(w, errors.New(http.StatusText(status)), status)
			return
		}
	})
}

func (a *App) canDoEverything(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, _ := a.sessionUser(r); u.CanDoEverything() {
			next.ServeHTTP(w, r)
		} else {
			status := http.StatusUnauthorized
			a.renderErrorPage(w, errors.New(http.StatusText(status)), status)
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
				a.renderErrorPage(w, fmt.Errorf("%s", err), http.StatusInternalServerError)
				return
			}
		}()

		next.ServeHTTP(w, r)
	})
}
