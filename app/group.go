package app

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mattfan00/jvbe/group"
)

func (a *App) renderNewGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := a.sessionUser(r)

		a.renderPage(w, "group/new.html", BaseData{
			User: u,
		})
	}
}

func (a *App) createGroup() http.HandlerFunc {
	type request struct {
		Name string `schema:"name"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := a.sessionUser(r)

		req, err := schemaDecode[request](r)
		if err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		_, err = a.groupService.CreateAndAddMember(group.CreateParams{
			CreatorId: u.Id,
			Name:      req.Name,
		})
		if err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/group/list", http.StatusSeeOther)
	}
}

func (a *App) renderGroupDetails() http.HandlerFunc {
	type data struct {
		BaseData
		Group group.GroupDetailed
	}

	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := a.sessionUser(r)
		id := chi.URLParam(r, "id")

		if !u.CanModifyGroup() {
			if err := a.groupService.UserCanAccessError(sql.NullString{
				String: id,
				Valid:  true,
			}, u.Id); err != nil {
				a.renderErrorPage(w, err, http.StatusInternalServerError)
				return
			}
		}

		g, err := a.groupService.GetDetailed(id)
		if err != nil {
			a.renderErrorPage(w, err, http.StatusInternalServerError)
			return
		}

		a.renderPage(w, "group/details.html", data{
			BaseData: BaseData{
				User: u,
			},
			Group: g,
		})
	}
}

func (a *App) renderGroupList() http.HandlerFunc {
	type data struct {
		BaseData
		Groups []group.Group
	}

	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := a.sessionUser(r)

		g, err := a.groupService.List()
		if err != nil {
			a.renderErrorPage(w, err, http.StatusInternalServerError)
			return
		}

		a.renderPage(w, "group/list.html", data{
			BaseData: BaseData{
				User: u,
			},
			Groups: g,
		})
	}
}

func (a *App) inviteGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := a.sessionUser(r)
		if !ok {
			a.session.Put(r.Context(), "redirect", r.URL.String())
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		id := chi.URLParam(r, "id")

		g, err := a.groupService.AddMemberFromInvite(id, u.Id)
		if err != nil {
			a.renderErrorPage(w, err, http.StatusInternalServerError)
			return
		}

		a.renderPage(w, "group/invite.html", g)
	}
}

func (a *App) renderEditGroup() http.HandlerFunc {
	type data struct {
		BaseData
		Group group.Group
	}

	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := a.sessionUser(r)
		id := chi.URLParam(r, "id")

		g, err := a.groupService.Get(id)
		if err != nil {
			a.renderErrorPage(w, err, http.StatusInternalServerError)
			return
		}

		a.renderPage(w, "group/edit.html", data{
			BaseData: BaseData{
				User: u,
			},
			Group: g,
		})
	}
}

func (a *App) updateGroup() http.HandlerFunc {
	type request struct {
		Name string `schema:"name"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		req, err := schemaDecode[request](r)
		if err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		err = a.groupService.Update(group.UpdateParams{
			Id:   id,
			Name: req.Name,
		})
		if err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/group/"+id, http.StatusSeeOther)
	}
}

func (a *App) deleteGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		err := a.groupService.Delete(id)
		if err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/group/list", http.StatusSeeOther)
	}
}

func (a *App) removeGroupMember() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		userId := chi.URLParam(r, "userId")

		err := a.groupService.RemoveMember(id, userId)
		if err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/group/"+id, http.StatusSeeOther)
	}
}

func (a *App) refreshInviteLinkGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		err := a.groupService.RefreshInviteId(id)
		if err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		w.Header().Add("HX-Location", "/group/"+id)
		w.Write(nil)
	}
}
