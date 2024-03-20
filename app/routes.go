package app

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/mattfan00/jvbe/auditlog"
)

func (a *App) Routes() http.Handler {
	r := chi.NewRouter()

	publicFileServer := http.FileServer(http.Dir("./ui/public"))
	r.Handle("/public/*", http.StripPrefix("/public/", publicFileServer))

	r.Get("/privacy", a.renderPrivacy())

	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitAll(100, 1*time.Minute))
		r.Use(middleware.Logger)
		r.Use(a.recoverPanic)
		r.Use(a.session.LoadAndSave)

		r.Get("/", a.renderIndex())

		r.Route("/auth", func(r chi.Router) {
			r.Get("/login", a.renderLogin())
			r.Get("/callback", a.handleLoginCallback())

			r.With(a.requireAuth).Get("/logout", a.handleLogout())
		})

		r.Group(func(r chi.Router) {
			r.Use(a.requireAuth)

			r.Get("/home", a.renderHome())
			r.With(a.canDoEverything).Get("/admin", a.renderAdmin())
			r.With(a.canDoEverything).Get("/auditlog", a.renderAuditlog())

			r.Route("/event", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(a.canModifyEvent)

					r.Get("/new", a.renderNewEvent())
					r.Post("/new", a.createEvent())
					r.Get("/{id}/edit", a.renderEditEvent())
					r.Post("/{id}/edit", a.updateEvent())
					r.Delete("/{id}/edit", a.deleteEvent())
				})

				r.Get("/{id}", a.renderEventDetails())
				r.Post("/respond", a.respondEvent())
			})
		})

		r.Route("/group", func(r chi.Router) {
			r.Get("/{id}/invite", a.inviteGroup())

			r.Group(func(r chi.Router) {
				r.Use(a.requireAuth)

				r.Group(func(r chi.Router) {
					r.Use(a.canModifyGroup)

					r.Get("/list", a.renderGroupList())
					r.Get("/new", a.renderNewGroup())
					r.Post("/new", a.createGroup())
					r.Get("/{id}/edit", a.renderEditGroup())
					r.Post("/{id}/edit", a.updateGroup())
					r.Delete("/{id}/edit", a.deleteGroup())
					r.Delete("/{id}/member/{userId}", a.removeGroupMember())
					r.Post("/{id}/invite", a.refreshInviteLinkGroup())
				})

				r.Get("/{id}", a.renderGroupDetails())
			})
		})

		r.Route("/review", func(r chi.Router) {
			r.Get("/request", a.renderReviewRequest())
			r.Post("/request", a.updateReview())

			r.Group(func(r chi.Router) {
				r.Use(a.canReviewUser)

				r.Get("/list", a.renderReviewList())
				r.Post("/approve", a.approveReview())
			})
		})

	})

	return r
}

func (a *App) renderPrivacy() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.renderPage(w, "privacy.html", nil)
	}
}

func (a *App) renderIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := a.sessionUser(r)
		if ok {
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}

		a.renderPage(w, "index.html", nil)
	}
}

func (a *App) renderAdmin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := a.sessionUser(r)

		a.renderPage(w, "admin.html", BaseData{
			User: u,
		})
	}
}

func (a *App) renderAuditlog() http.HandlerFunc {
	type data struct {
		BaseData
		AuditLogs []auditlog.AuditLog
		CurrPage  int
		MaxPage   int
		PrevPage  int
		NextPage  int
	}

	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := a.sessionUser(r)

		pageQuery := r.URL.Query().Get("page")
		var page = 1
		if pageQuery != "" {
			page, _ = strconv.Atoi(pageQuery)
		}
		pageSize := 10

		al, count, err := a.auditlogService.List(auditlog.ListFilter{
			Limit:  pageSize,
			Offset: pageSize * (page - 1),
		})
		if err != nil {
			a.renderErrorPage(w, err, http.StatusInternalServerError)
			return
		}

		prevPage := page - 1
		if prevPage < 1 {
			prevPage = 1
		}

		maxPage := int(math.Ceil(float64(count) / float64(pageSize)))
		nextPage := page + 1
		if nextPage > maxPage {
			nextPage = maxPage
		}

		a.renderPage(w, "auditlog.html", data{
			BaseData: BaseData{
				User: u,
			},
			AuditLogs: al,
			CurrPage:  page,
			MaxPage:   maxPage,
			PrevPage:  prevPage,
			NextPage:  nextPage,
		})
	}
}
