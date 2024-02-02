package app

import (
	eventPkg "github/mattfan00/jvbe/event"
	userPkg "github/mattfan00/jvbe/user"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
)

func (a *App) Routes() http.Handler {
	r := chi.NewRouter()

	publicFileServer := http.FileServer(http.Dir("./ui/public"))
	r.Handle("/public/*", http.StripPrefix("/public/", publicFileServer))

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

				r.Group(func(r chi.Router) {
					r.Use(a.requireAdmin)

					r.Get("/new", a.renderNewEvent)
					r.Post("/", a.createEvent)
				})
			})
		})
	})

	return r
}

func (a *App) renderIndex(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.getSessionUser(r); ok {
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
	u, _ := a.session.Get(r.Context(), "user").(userPkg.SessionUser)

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
	u, _ := a.session.Get(r.Context(), "user").(userPkg.SessionUser)

	a.renderPage(w, "event/new.html", BaseData{
		User: u,
	})
}

type eventDetailsData struct {
	BaseData
	Event eventPkg.EventDetailed
}

func (a *App) renderEventDetails(w http.ResponseWriter, r *http.Request) {
	u, _ := a.session.Get(r.Context(), "user").(userPkg.SessionUser)
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
		Event: e,
	})
}

func (a *App) respondEvent(w http.ResponseWriter, r *http.Request) {
	u, _ := a.session.Get(r.Context(), "user").(userPkg.SessionUser)

	err := r.ParseForm()
	if err != nil {
		a.renderErrorNotif(w, err, http.StatusInternalServerError)
		return
	}

	var req eventPkg.RespondEventRequest
	err = schema.NewDecoder().Decode(&req, r.PostForm)
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

	err = a.event.CreateFromRequest(req)
	if err != nil {
		a.renderErrorNotif(w, err, http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func (a *App) renderLogin(w http.ResponseWriter, r *http.Request) {
	url := a.auth.GetOauthLoginUrl()
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (a *App) handleLoginCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	err := a.auth.ValidateState(state)
	if err != nil {
		a.renderErrorPage(w, err, http.StatusInternalServerError)
		return
	}

	// now that we are succesfully authenticated, use the code we got back to get the access token
	code := r.URL.Query().Get("code")
	u, err := a.auth.GetUserFromOauthCode(code)
	if err != nil {
		a.renderErrorPage(w, err, http.StatusInternalServerError)
		return
	}

	sessionUser := u.ToSessionUser()

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

	a.session.Remove(r.Context(), "user")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
