package app

import (
	"github/mattfan00/jvbe/auth"
	"github/mattfan00/jvbe/event"
	"github/mattfan00/jvbe/template"
	"github/mattfan00/jvbe/user"

	"github.com/alexedwards/scs/v2"
)

type App struct {
	event *event.Service
	user  *user.Service
	auth  *auth.Service

	session   *scs.SessionManager
	templates template.Map
}

func New(
	event *event.Service,
	user *user.Service,
	auth *auth.Service,

	session *scs.SessionManager,
	templates template.Map,
) *App {
	return &App{
		event: event,
		user:  user,
		auth:  auth,

		session:   session,
		templates: templates,
	}
}
