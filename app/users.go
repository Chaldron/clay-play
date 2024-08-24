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
