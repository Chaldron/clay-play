package app

import (
	"bytes"
	"github.com/mattfan00/jvbe/auth"
	"github.com/mattfan00/jvbe/config"
	"github.com/mattfan00/jvbe/event"
	"github.com/mattfan00/jvbe/group"
	"github.com/mattfan00/jvbe/template"
	"github.com/mattfan00/jvbe/user"
	"log"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/schema"
)

type App struct {
	eventService event.Service
	userService  user.Service
	authService  auth.Service
	groupService group.Service

	conf      *config.Config
	session   *scs.SessionManager
	templates template.TemplateMap
}

func New(
	eventService event.Service,
	userService user.Service,
	authService auth.Service,
	groupService group.Service,

	conf *config.Config,
	session *scs.SessionManager,
	templates template.TemplateMap,
) *App {
	return &App{
		eventService: eventService,
		userService:  userService,
		authService:  authService,
		groupService: groupService,

		conf:      conf,
		session:   session,
		templates: templates,
	}
}

type BaseData struct {
	User user.SessionUser
}

func (a *App) renderTemplate(
	w http.ResponseWriter,
	template string,
	templateName string,
	data any,
) {
	t, ok := a.templates[template]
	if !ok {
		http.Error(w, "template not found", http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)

	err := t.ExecuteTemplate(buf, templateName, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf.WriteTo(w)
}

func (a *App) renderPage(
	w http.ResponseWriter,
	template string,
	data any,
) {
	a.renderTemplate(w, template, "base", data)
}

func errorLog(err error) {
	log.Printf("ERROR: %s", err.Error())
}

func (a *App) renderErrorNotif(
	w http.ResponseWriter,
	err error,
	status int,
) {
	errorLog(err)
	w.Header().Add("HX-Reswap", "none") // so that UI does not swap rest of the blank template
	w.WriteHeader(status)
	a.renderTemplate(w, "error-notif.html", "error", map[string]any{
		"Error": err,
	})
}

func (a *App) renderErrorPage(
	w http.ResponseWriter,
	err error,
	status int,
) {
	errorLog(err)
	w.Header().Add("HX-Retarget", "body")
	w.Header().Add("HX-Reswap", "innerHTML")
	w.WriteHeader(status)
	a.renderPage(w, "error.html", map[string]any{
		"Error": err,
	})
}

func (a *App) renewSessionUser(r *http.Request, u *user.SessionUser) error {
	err := a.session.RenewToken(r.Context())
	if err != nil {
		return err
	}

	a.session.Put(r.Context(), "user", u)

	return nil
}

func schemaDecode[T any](r *http.Request) (T, error) {
	var v T

	if err := r.ParseForm(); err != nil {
		return v, err
	}

	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	if err := decoder.Decode(&v, r.PostForm); err != nil {
		return v, err
	}

	return v, nil
}
