package app

import (
	"errors"
	"fmt"
	eventPkg "github/mattfan00/jvbe/event"
	groupPkg "github/mattfan00/jvbe/group"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/schema"
)

func (a *App) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	publicFileServer := http.FileServer(http.Dir("./ui/public"))
	r.Handle("/public/*", http.StripPrefix("/public/", publicFileServer))

	r.Get("/privacy", a.renderPrivacy)

	r.Group(func(r chi.Router) {
		r.Use(a.recoverPanic)
		r.Use(a.session.LoadAndSave)

		r.Get("/", a.renderIndex)

		r.Route("/auth", func(r chi.Router) {
			r.Get("/login", a.renderLogin)
			r.Get("/callback", a.handleLoginCallback)

			r.With(a.requireAuth).Get("/logout", a.handleLogout)
		})

		r.Group(func(r chi.Router) {
			r.Use(a.requireAuth)

			r.Get("/home", a.renderHome)

			r.Route("/event", func(r chi.Router) {
				r.Get("/{id}", a.renderEventDetails)
				r.Post("/respond", a.respondEvent)

				r.With(a.canCreateEvent).Get("/new", a.renderNewEvent)
				r.With(a.canCreateEvent).Post("/", a.createEvent)
				r.With(a.canDeleteEvent).Delete("/{id}", a.deleteEvent)
			})

			r.Route("/group", func(r chi.Router) {
				r.Get("/list", a.renderGroupList)
				r.Get("/new", a.renderNewGroup)
				r.Get("/{id}", a.renderGroupDetails)
				r.Get("/{id}/invite", a.inviteGroup)
				r.Post("/", a.createGroup)
				r.Get("/{id}/edit", a.renderEditGroup)
				r.Put("/{id}", a.updateGroup)
				r.Delete("/{id}", a.deleteGroup)
				r.Delete("/{id}/member/{userId}", a.removeGroupMember)
			})
		})
	})

	return r
}

func (a *App) renderPrivacy(w http.ResponseWriter, r *http.Request) {
	a.renderPage(w, "privacy.html", nil)
}

func (a *App) renderIndex(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.sessionUser(r); ok {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	a.renderPage(w, "index.html", nil)
}

type homeData struct {
	BaseData
	CurrEvents []eventPkg.Event
}

func (a *App) renderHome(w http.ResponseWriter, r *http.Request) {
	u, _ := a.sessionUser(r)

	currEvents, err := a.event.GetCurrent(u.Id)
	if err != nil {
		a.renderErrorPage(w, err, http.StatusInternalServerError)
		return
	}

	a.renderPage(w, "home.html", homeData{
		BaseData: BaseData{
			User: u,
		},
		CurrEvents: currEvents,
	})
}

func (a *App) renderNewEvent(w http.ResponseWriter, r *http.Request) {
	u, _ := a.sessionUser(r)

	a.renderPage(w, "event/new.html", BaseData{
		User: u,
	})
}

type eventDetailsData struct {
	BaseData
	Event            eventPkg.EventDetailed
	MaxAttendeeCount int
}

func (a *App) renderEventDetails(w http.ResponseWriter, r *http.Request) {
	u, _ := a.sessionUser(r)
	eventId := chi.URLParam(r, "id")

	e, err := a.event.GetDetailed(eventId, u.Id)
	if err != nil {
		a.renderErrorPage(w, err, http.StatusInternalServerError)
		return
	}

	a.renderPage(w, "event/details.html", eventDetailsData{
		BaseData: BaseData{
			User: u,
		},
		Event:            e,
		MaxAttendeeCount: eventPkg.MaxAttendeeCount,
	})
}

func (a *App) respondEvent(w http.ResponseWriter, r *http.Request) {
	u, _ := a.sessionUser(r)

	err := r.ParseForm()
	if err != nil {
		a.renderErrorNotif(w, err, http.StatusInternalServerError)
		return
	}

	var req eventPkg.RespondEventRequest
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	err = decoder.Decode(&req, r.PostForm)
	if err != nil {
		a.renderErrorNotif(w, err, http.StatusInternalServerError)
		return
	}

	err = a.event.HandleEventResponse(u.Id, req)
	if err != nil {
		a.renderErrorNotif(w, err, http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/event/"+req.Id, http.StatusSeeOther)
}

func (a *App) createEvent(w http.ResponseWriter, r *http.Request) {
	u, _ := a.sessionUser(r)

	err := r.ParseForm()
	if err != nil {
		a.renderErrorNotif(w, err, http.StatusInternalServerError)
		return
	}

	var req eventPkg.CreateEventRequest
	err = schema.NewDecoder().Decode(&req, r.PostForm)
	if err != nil {
		a.renderErrorNotif(w, err, http.StatusInternalServerError)
		return
	}
    req.Creator = u.FirstName

	err = a.event.CreateFromRequest(req)
	if err != nil {
		a.renderErrorNotif(w, err, http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func (a *App) deleteEvent(w http.ResponseWriter, r *http.Request) {
	eventId := chi.URLParam(r, "id")
	err := a.event.Delete(eventId)
	if err != nil {
		a.renderErrorNotif(w, err, http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

var expectedStateVal = "hellothisisalongstate"

func (a *App) renderLogin(w http.ResponseWriter, r *http.Request) {
	// TODO: state should be random
	url := a.auth.AuthCodeUrl(expectedStateVal)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

type customClaims struct {
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

func (a *App) handleLoginCallback(w http.ResponseWriter, r *http.Request) {
	log.Printf("callback: %s", r.URL.String())

	state := r.URL.Query().Get("state")
	if state != expectedStateVal {
		err := fmt.Errorf("invalid oauth state, expected '%s', got '%s'", expectedStateVal, state)
		a.renderErrorPage(w, err, http.StatusInternalServerError)
		return
	}

	// now that we are succesfully authenticated, use the code we got back to get the access token
	code := r.URL.Query().Get("code")
	externalUser, accessToken, err := a.auth.InfoFromProvider(code)
	if err != nil {
		a.renderErrorPage(w, err, http.StatusInternalServerError)
		return
	}

	u, err := a.user.HandleFromExternal(externalUser)
	if err != nil {
		a.renderErrorPage(w, err, http.StatusInternalServerError)
		return
	}

	sessionUser := u.ToSessionUser()

	// dont verify token since this should have come from oauth and I don't want to deal with verifying right now
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256"}))
	token, _, err := parser.ParseUnverified(accessToken, &customClaims{})
	if err != nil {
		a.renderErrorPage(w, err, http.StatusInternalServerError)
		return
	}

	claims, ok := token.Claims.(*customClaims)
	if !ok {
		a.renderErrorPage(w, errors.New("cannot get claims"), http.StatusInternalServerError)
		return
	}

	sessionUser.Permissions = claims.Permissions

	err = a.session.RenewToken(r.Context())
	if err != nil {
		a.renderErrorPage(w, err, http.StatusInternalServerError)
		return
	}

	a.session.Put(r.Context(), "user", &sessionUser)

	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func (a *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	err := a.session.Destroy(r.Context())
	if err != nil {
		a.renderErrorPage(w, err, http.StatusInternalServerError)
		return
	}

	redirect := fmt.Sprintf("https://%s/logout?redirect=%s", a.conf.Oauth.Domain, a.conf.OauthLogoutRedirectUrl())
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func (a *App) renderNewGroup(w http.ResponseWriter, r *http.Request) {
	u, _ := a.sessionUser(r)

	a.renderPage(w, "group/new.html", BaseData{
		User: u,
	})
}

func (a *App) createGroup(w http.ResponseWriter, r *http.Request) {
	u, _ := a.sessionUser(r)

	err := r.ParseForm()
	if err != nil {
		a.renderErrorNotif(w, err, http.StatusInternalServerError)
		return
	}

	var req groupPkg.CreateRequest
	err = schema.NewDecoder().Decode(&req, r.PostForm)
	if err != nil {
		a.renderErrorNotif(w, err, http.StatusInternalServerError)
		return
	}
	req.CreatorId = u.Id

	err = a.group.CreateAndAddMember(req)
	if err != nil {
		a.renderErrorNotif(w, err, http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/group/list", http.StatusSeeOther)
}

type groupDetailsData struct {
	BaseData
	Group groupPkg.GroupDetailed
}

func (a *App) renderGroupDetails(w http.ResponseWriter, r *http.Request) {
	u, _ := a.sessionUser(r)
	id := chi.URLParam(r, "id")

	g, err := a.group.GetDetailed(id)
	if err != nil {
		a.renderErrorPage(w, err, http.StatusInternalServerError)
		return
	}

	a.renderPage(w, "group/details.html", groupDetailsData{
		BaseData: BaseData{
			User: u,
		},
		Group: g,
	})
}

type groupListData struct {
	BaseData
	Groups []groupPkg.Group
}

func (a *App) renderGroupList(w http.ResponseWriter, r *http.Request) {
	u, _ := a.sessionUser(r)

	g, err := a.group.List()
	if err != nil {
		a.renderErrorPage(w, err, http.StatusInternalServerError)
		return
	}

	a.renderPage(w, "group/list.html", groupListData{
		BaseData: BaseData{
			User: u,
		},
		Groups: g,
	})
}

func (a *App) inviteGroup(w http.ResponseWriter, r *http.Request) {
	u, ok := a.sessionUser(r)
	if !ok {
		w.Write([]byte("need to redirect login here"))
		return
	}

	id := chi.URLParam(r, "id")

	g, err := a.group.AddMemberFromInvite(id, u.Id)
	if err != nil {
		a.renderErrorPage(w, err, http.StatusInternalServerError)
		return
	}

	a.renderPage(w, "group/invite.html", g)
}

type editGroupData struct {
	BaseData
	Group groupPkg.Group
}

func (a *App) renderEditGroup(w http.ResponseWriter, r *http.Request) {
	u, _ := a.sessionUser(r)
	id := chi.URLParam(r, "id")

	g, err := a.group.Get(id)
	if err != nil {
		a.renderErrorPage(w, err, http.StatusInternalServerError)
		return
	}

	a.renderPage(w, "group/edit.html", editGroupData{
		BaseData: BaseData{
			User: u,
		},
		Group: g,
	})
}

func (a *App) updateGroup(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := r.ParseForm()
	if err != nil {
		a.renderErrorNotif(w, err, http.StatusInternalServerError)
		return
	}

	var req groupPkg.UpdateRequest
	err = schema.NewDecoder().Decode(&req, r.PostForm)
	if err != nil {
		a.renderErrorNotif(w, err, http.StatusInternalServerError)
		return
	}
	req.Id = id

	err = a.group.Update(req)
	if err != nil {
		a.renderErrorNotif(w, err, http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/group/"+id, http.StatusSeeOther)
}

func (a *App) deleteGroup(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := a.group.Delete(id)
	if err != nil {
		a.renderErrorNotif(w, err, http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/group/list", http.StatusSeeOther)
}

func (a *App) removeGroupMember(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userId := chi.URLParam(r, "userId")

	err := a.group.RemoveMember(id, userId)
	if err != nil {
		a.renderErrorNotif(w, err, http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/group/"+id, http.StatusSeeOther)
}
