package event

import (
	"github/mattfan00/jvbe/template"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
)

type Service struct {
	store     *Store
	templates template.Map
}

func NewService(store *Store, templates template.Map) *Service {
	return &Service{
		store:     store,
		templates: templates,
	}
}

func (s *Service) Routes(r *chi.Mux) {
	r.Get("/", s.getIndexPage)
	r.Route("/event", func(r chi.Router) {
		r.Get("/new", s.getNewEventPage)
		r.Get("/{id}", s.getEventPage)
		r.Post("/", s.postEventHandler)
	})
}

func (s *Service) getIndexPage(w http.ResponseWriter, r *http.Request) {
	currEvents, err := s.store.GetCurrent()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.templates["home.html"].ExecuteTemplate(w, "base", map[string]any{
		"CurrEvents": currEvents,
	})
}

func (s *Service) getNewEventPage(w http.ResponseWriter, r *http.Request) {
	s.templates["event/new.html"].ExecuteTemplate(w, "base", nil)
}

func (s *Service) getEventPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hi"))
}

type eventRequest struct {
	Name           string `schema:"name"`
	Capacity       int    `schema:"capacity"`
	Start          string `schema:"start"`
	TimezoneOffset int    `schema:"timezoneOffset"`
	Location       string `schema:"location"`
}

func (s *Service) postEventHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req eventRequest
	err = schema.NewDecoder().Decode(&req, r.PostForm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	start, err := time.Parse("2006-01-02T15:04", req.Start)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	start = start.Add(time.Minute * time.Duration(req.TimezoneOffset))

	newEvent := Event{
		Name:      req.Name,
		Capacity:  req.Capacity,
		Start:     start,
		Location:  req.Location,
		CreatedAt: time.Now(),
	}

	err = s.store.InsertOne(newEvent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
