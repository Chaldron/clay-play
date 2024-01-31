package event

import (
	"github/mattfan00/jvbe/template"
	"sync"
	"time"
)

type Service struct {
	store             *Store
	templates         template.Map
	eventResponseLock sync.Mutex
}

func NewService(store *Store, templates template.Map) *Service {
	return &Service{
		store:     store,
		templates: templates,
	}
}

func (s *Service) GetCurrent(userId string) ([]Event, error) {
	currEvents, err := s.store.GetCurrent(userId)
	if err != nil {
		return []Event{}, err
	}

	return currEvents, nil
}

func (s *Service) CreateFromRequest(req CreateEventRequest) error {
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

func (s *Service) HandleEventResponse(userId string, req RespondEventRequest) (Event, error) {
	s.eventResponseLock.Lock()
	defer s.eventResponseLock.Unlock()

	e := EventResponse{
		EventId: req.Id,
		UserId:  userId,
		Going:   req.Going,
	}

	event, err := s.store.UpdateResponse(e)
	if err != nil {
		return Event{}, err
	}

	return event, nil
}
