package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/mattfan00/jvbe/user"
	"net/http"
	"strconv"
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

func (a *App) renderEditUser() http.HandlerFunc {
	type data struct {
		BaseData
		UserData user.User
	}

	return func(w http.ResponseWriter, r *http.Request) {
		su, _ := a.sessionUser(r)

		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			a.renderErrorPage(w, err, http.StatusBadRequest)
			return
		}

		u, err := a.userService.Get(id)
		// Convert the id from string to int64

		if err != nil {
			a.renderErrorPage(w, err, http.StatusInternalServerError)
			return
		}

		a.renderPage(w, "user/edit.html", data{
			BaseData: BaseData{
				User: su,
			},
			UserData: u,
		})
	}
}

func (a *App) updateUser() http.HandlerFunc {
	type request struct {
		Name     string `schema:"name"`
		Email    string `schema:"email"`
		Password string `schema:"password"`
		IsAdmin  bool   `schema:"isadmin"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			a.renderErrorPage(w, err, http.StatusBadRequest)
			return
		}

		a.auditlogService.Create(id, "Edited")

		req, err := schemaDecode[request](r)
		if err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		_, err = a.userService.Update(user.UpdateParams{
			Id:       id,
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

func (a *App) deleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			a.renderErrorPage(w, err, http.StatusBadRequest)
			return
		}

		err = a.userService.Delete(id)
		if err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/user/list", http.StatusSeeOther)
	}
}
