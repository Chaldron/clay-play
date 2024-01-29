package event

import (
	"github/mattfan00/jvbe/template"
	"time"
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

func (s *Service) GetCurrent() ([]Event, error) {
	currEvents, err := s.store.GetCurrent()
	if err != nil {
		return []Event{}, err
	}

	return currEvents, nil
}

func (s *Service) CreateFromRequest(req EventRequest) error {
	start, err := time.Parse("2006-01-02T15:04", req.Start)
	if err != nil {
		return err
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
		return err
	}

	return nil
}
