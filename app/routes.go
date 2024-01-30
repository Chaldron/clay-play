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
		r.Use(a.session.LoadAndSave)

		r.Get("/", a.RenderIndex)

		r.Route("/event", func(r chi.Router) {
			r.Get("/new", a.RenderNewEvent)
			r.Get("/{id}", a.RenderSingleEvent)
			r.Post("/", a.CreateEvent)
		})

		r.Route("/auth", func(r chi.Router) {
			r.Get("/login", a.RenderLogin)
			r.Get("/callback", a.HandleLoginCallback)
			r.Get("/logout", a.HandleLogout)
		})
	})

	return r
}

type indexData struct {
	BaseData
	CurrEvents []eventPkg.Event
}

func (a *App) RenderIndex(w http.ResponseWriter, r *http.Request) {
	u, _ := a.session.Get(r.Context(), "user").(userPkg.SessionUser)

	currEvents, err := a.event.GetCurrent()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.templates["home.html"].ExecuteTemplate(w, "base", indexData{
		BaseData: BaseData{
			User: u,
		},
		CurrEvents: currEvents,
	})
}

func (a *App) RenderNewEvent(w http.ResponseWriter, r *http.Request) {
	a.templates["event/new.html"].ExecuteTemplate(w, "base", nil)
}

func (a *App) RenderSingleEvent(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hi"))
}

func (a *App) CreateEvent(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req eventPkg.EventRequest
	err = schema.NewDecoder().Decode(&req, r.PostForm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = a.event.CreateFromRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) RenderLogin(w http.ResponseWriter, r *http.Request) {
	url := a.auth.GetOauthLoginUrl()
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (a *App) HandleLoginCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	err := a.auth.ValidateState(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// now that we are succesfully authenticated, use the code we got back to get the access token
	code := r.URL.Query().Get("code")
	u, err := a.auth.GetUserFromOauthCode(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sessionUser := u.ToSessionUser()

	err = a.session.RenewToken(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.session.Put(r.Context(), "user", &sessionUser)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) HandleLogout(w http.ResponseWriter, r *http.Request) {
	err := a.session.RenewToken(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	a.session.Remove(r.Context(), "user")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
