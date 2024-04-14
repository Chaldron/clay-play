package app

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
	"github.com/mattfan00/jvbe/event"
	"github.com/mattfan00/jvbe/group"
	"github.com/mattfan00/jvbe/template"
)

func (a *App) renderHome() http.HandlerFunc {
	type data struct {
		BaseData
		CurrEvents []event.Event
		PastEvents []event.Event
	}

	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := a.sessionUser(r)

		currEvents, err := a.eventService.List(event.ListFilter{
			Upcoming: true,
			UserId:   u.Id,
		})
		if err != nil {
			a.renderErrorPage(w, err, http.StatusInternalServerError)
			return
		}

		pastEvents, err := a.eventService.List(event.ListFilter{
			Past:        true,
			OrderByDesc: true,
			Limit:       10,
			UserId:      u.Id,
		})
		if err != nil {
			a.renderErrorPage(w, err, http.StatusInternalServerError)
			return
		}

		a.renderPage(w, "home.html", data{
			BaseData: BaseData{
				User: u,
			},
			CurrEvents: currEvents.Events,
			PastEvents: pastEvents.Events,
		})
	}
}

func (a *App) renderNewEvent() http.HandlerFunc {
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

		a.renderPage(w, "event/new.html", data{
			BaseData: BaseData{
				User: u,
			},
			Groups: g,
		})
	}
}

func (a *App) createEvent() http.HandlerFunc {
	type request struct {
		Name           string `schema:"name"`
		GroupId        string `schema:"groupId"`
		Capacity       int    `schema:"capacity"`
		Start          string `schema:"start"`
		TimezoneOffset int    `schema:"timezoneOffset"`
		Location       string `schema:"location"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := a.sessionUser(r)

		req, err := schemaDecode[request](r)
		if err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		start, err := timeFromForm(req.Start, req.TimezoneOffset)
		if err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		_, err = a.eventService.Create(event.CreateParams{
			Name:      req.Name,
			GroupId:   req.GroupId,
			Capacity:  req.Capacity,
			Start:     start,
			Location:  req.Location,
			CreatorId: u.Id,
		})
		if err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/home", http.StatusSeeOther)
	}
}

func (a *App) renderEditEvent() http.HandlerFunc {
	type data struct {
		BaseData
		Event event.Event
	}

	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := a.sessionUser(r)
		id := chi.URLParam(r, "id")

		e, err := a.eventService.Get(id)
		if err != nil {
			a.renderErrorPage(w, err, http.StatusInternalServerError)
			return
		}

		a.renderPage(w, "event/edit.html", data{
			BaseData: BaseData{
				User: u,
			},
			Event: e,
		})
	}
}

func (a *App) updateEvent() http.HandlerFunc {
	type request struct {
		Name           string `schema:"name"`
		Capacity       int    `schema:"capacity"`
		Start          string `schema:"start"`
		TimezoneOffset int    `schema:"timezoneOffset"`
		Location       string `schema:"location"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := a.sessionUser(r)
		id := chi.URLParam(r, "id")
		a.log.Printf("user updating event %s: %s", id, u.Id)

		if err := r.ParseForm(); err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		var req request
		if err := schema.NewDecoder().Decode(&req, r.PostForm); err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		start, err := timeFromForm(req.Start, req.TimezoneOffset)
		if err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		if err := a.eventService.Update(event.UpdateParams{
			Id:   id,
			Name: req.Name,
			Capacity: req.Capacity,
			Start:    start,
			Location: req.Location,
		}); err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/event/"+id, http.StatusSeeOther)
		w.Write(nil)
	}
}

func (a *App) deleteEvent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		err := a.eventService.Delete(id)
		if err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/home", http.StatusSeeOther)
	}
}

func (a *App) renderEventDetails() http.HandlerFunc {
	type data struct {
		BaseData
		Event            event.EventDetailed
		MaxAttendeeCount int
	}

	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := a.sessionUser(r)
		id := chi.URLParam(r, "id")

		e, err := a.eventService.GetDetailed(id, u.Id)
		if err != nil {
			a.renderErrorPage(w, err, http.StatusInternalServerError)
			return
		}

		if err = a.groupService.UserCanAccessError(e.GroupId, u.Id); err != nil {
			a.renderErrorPage(w, err, http.StatusInternalServerError)
			return
		}

		a.renderPage(w, "event/details.html", data{
			BaseData: BaseData{
				User: u,
			},
			Event:            e,
			MaxAttendeeCount: event.MaxAttendeeCount,
		})
	}
}

func (a *App) respondEvent() http.HandlerFunc {
	type request struct {
		Id            string `schema:"id"`
		AttendeeCount int    `schema:"attendeeCount"`
	}
	var lock sync.Mutex

	return func(w http.ResponseWriter, r *http.Request) {
		lock.Lock()
		defer lock.Unlock()

		u, _ := a.sessionUser(r)

		req, err := schemaDecode[request](r)
		if err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		e, err := a.eventService.Get(req.Id)
		if err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		if err = a.groupService.UserCanAccessError(e.GroupId, u.Id); err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		err = a.eventService.HandleResponse(event.HandleResponseParams{
			UserId:        u.Id,
			Id:            req.Id,
			AttendeeCount: req.AttendeeCount,
		})
		if err != nil {
			a.renderErrorNotif(w, err, http.StatusInternalServerError)
			return
		}

		err = a.auditlogService.Create(
			u.Id,
			fmt.Sprintf("Responded to <a href=\"/event/%s\">%s</a> with %d attendee(s)", e.Id, e.Name, req.AttendeeCount),
		)
		if err != nil {
			a.log.Errorf(err.Error())
		}

		http.Redirect(w, r, "/event/"+req.Id, http.StatusSeeOther)
	}
}

func timeFromForm(t string, offset int) (time.Time, error) {
	r, err := time.Parse(template.FormTimeFormat, t)
	if err != nil {
		return time.Time{}, err
	}
	r = r.Add(time.Minute * time.Duration(offset))

	return r, nil
}
