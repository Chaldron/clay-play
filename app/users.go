package app

import (
	"github.com/mattfan00/jvbe/user"
	"net/http"
)

func (a *App) renderUserList() http.HandlerFunc {
	type data struct {
		BaseData
		Users []user.User
	}

	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := a.sessionUser(r)

		u_list, err := a.userService.GetAll()
		if err != nil {
			a.renderErrorPage(w, err, http.StatusInternalServerError)
			return
		}

		a.renderPage(w, "user/list.html", data{
			BaseData: BaseData{
				User: u,
			},
			Users: u_list,
		})
	}
}

func (a *App) renderNewUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := a.sessionUser(r)

		a.renderPage(w, "user/new.html", BaseData{
			User: u,
		})
	}
}

func (a *App) createUser() http.HandlerFunc {
	type request struct {
		Name     string `schema:"name"`
		Email    string `schema:"email"`
		Password string `schema:"password"`
		IsAdmin  bool   `schema:"isadmin"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req, err := schemaDecode[request](r)
		if err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		_, err = a.userService.Create(user.CreateParams{
			FullName: req.Name,
			Email:    req.Email,
			Password: req.Password,
			IsAdmin:  req.IsAdmin,
		})
		if err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/user/list", http.StatusSeeOther)
	}
}
