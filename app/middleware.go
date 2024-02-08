package app

import (
	"errors"
	"fmt"
	user "github/mattfan00/jvbe/user"
	"net/http"

	_ "github.com/golang-jwt/jwt/v5"
)

func (a *App) sessionUser(r *http.Request) (user.SessionUser, bool) {
	u, ok := a.session.Get(r.Context(), "user").(user.SessionUser)
	o := ok && u.IsAuthenticated()
	return u, o
}

func (a *App) accessToken(r *http.Request) (string, bool) {
	at := a.session.GetString(r.Context(), "accessToken")
	return at, at != ""
}

// TODO: clear sessions if auth fails
func (a *App) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, sessionUserOk := a.sessionUser(r)
		_, accessTokenOk := a.accessToken(r)

		/*
		   token, _ := jwt.Parse(accessToken, func(t *jwt.Token) (interface{}, error) { return "", nil })
		       if err != nil {
		           a.renderErrorPage(w, err, http.StatusForbidden)
		           return
		       }

		   fmt.Printf("%+v\n", token.Claims)
		*/

		if sessionUserOk && accessTokenOk {
			next.ServeHTTP(w, r)
		} else {
			status := http.StatusForbidden
			a.renderErrorPage(w, errors.New(http.StatusText(status)), status)
			return
		}
	})
}

func (a *App) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, _ := a.sessionUser(r); u.IsAdmin {
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
