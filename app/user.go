package app

import (
	"log"
	"net/http"

	"github.com/mattfan00/jvbe/user"
)

func (a *App) renderReviewRequest() http.HandlerFunc {
	type data struct {
		BaseData
		UserReview user.UserReview
	}

	return func(w http.ResponseWriter, r *http.Request) {
		su, _ := a.sessionUser(r)

		// recheck if the user is active so that user is redirected to application once they are
		u, err := a.userService.Get(su.Id)
		if err != nil {
			a.renderErrorPage(w, err, http.StatusInternalServerError)
			return
		}
		t := u.ToSessionUser()
		t.Permissions = su.Permissions
		su = t
		log.Printf("user at review: %+v", su)

		if err := a.renewSessionUser(r, &su); err != nil {
			a.renderErrorPage(w, err, http.StatusInternalServerError)
			return
		}

		if su.Status == user.UserStatusActive {
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}

		userReview, err := a.userService.GetReview(su.Id)
		if err != nil {
			a.renderErrorPage(w, err, http.StatusInternalServerError)
			return
		}

		a.renderPage(w, "review/request.html", data{
			BaseData: BaseData{
				User: su,
			},
			UserReview: userReview,
		})
	}
}

func (a *App) updateReview() http.HandlerFunc {
	type request struct {
		Comment string `schema:"comment"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := a.sessionUser(r)

		req, err := schemaDecode[request](r)
		if err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		err = a.userService.UpdateReview(user.UpdateReviewParams{
			UserId:  u.Id,
			Comment: req.Comment,
		})
		if err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		w.Write([]byte("Successfully updated your review!"))
	}
}

func (a *App) renderReviewList() http.HandlerFunc {
	type data struct {
		BaseData
		Reviews []user.UserReview
	}

	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := a.sessionUser(r)

		urs, err := a.userService.ListReviews()
		if err != nil {
			a.renderErrorPage(w, err, http.StatusInternalServerError)
			return
		}

		a.renderPage(w, "review/list.html", data{
			BaseData: BaseData{
				User: u,
			},
			Reviews: urs,
		})
	}
}

func (a *App) approveReview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := a.userService.ApproveReview(r.FormValue("user_id"))
		if err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/review/list", http.StatusSeeOther)
	}
}
