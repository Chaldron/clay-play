package app

import (
	"bytes"
	"github/mattfan00/jvbe/auth"
	"github/mattfan00/jvbe/config"
	"github/mattfan00/jvbe/event"
	"github/mattfan00/jvbe/group"
	"github/mattfan00/jvbe/template"
	"github/mattfan00/jvbe/user"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

type App struct {
	event *event.Service
	user  *user.Service
	auth  *auth.Service
	group *group.Service

	conf      *config.Config
	session   *scs.SessionManager
	templates template.TemplateMap
}

func New(
	event *event.Service,
	user *user.Service,
	auth *auth.Service,
	group *group.Service,

	conf *config.Config,
	session *scs.SessionManager,
	templates template.TemplateMap,
) *App {
	return &App{
		event: event,
		user:  user,
		auth:  auth,
		group: group,

		conf:      conf,
		session:   session,
		templates: templates,
	}
}

type BaseData struct {
	User  user.SessionUser
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

func (a *App) renderErrorNotif(
	w http.ResponseWriter,
	err error,
	status int,
) {
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
	w.Header().Add("HX-Retarget", "body")
	w.Header().Add("HX-Reswap", "innerHTML")
	w.WriteHeader(status)
	a.renderPage(w, "error.html", map[string]any{
		"Error": err,
	})
}
