package event

import (
	"errors"
	"fmt"
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

func (s *Service) GetDetailed(eventId string, userId string) (EventDetailed, error) {
	event, err := s.store.GetById(eventId)
	if err != nil {
		return EventDetailed{}, err
	}

	responses, err := s.store.GetResponsesByEventId(eventId)
	if err != nil {
		return EventDetailed{}, err
	}

	userResponse, err := s.store.GetUserResponse(eventId, userId)
	if err != nil {
		return EventDetailed{}, err
	}

	e := EventDetailed{
		Event:        event,
		UserResponse: userResponse,
		Responses:    responses,
	}

	return e, nil
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

func (s *Service) Delete(eventId string) error {
	err := s.store.DeleteById(eventId)
	if err != nil {
		return err
	}

	return nil
}

var MaxAttendeeCount = 2

func (s *Service) HandleEventResponse(userId string, req RespondEventRequest) error {
	s.eventResponseLock.Lock()
	defer s.eventResponseLock.Unlock()

	if req.AttendeeCount < 0 {
		return errors.New("cannot have less than 0 attendees")
	}

	if req.AttendeeCount > MaxAttendeeCount {
		return fmt.Errorf("maximum of %d plus one(s) allowed", MaxAttendeeCount - 1)
	}

	e := EventResponse{
		EventId:       req.Id,
		UserId:        userId,
		AttendeeCount: req.AttendeeCount,
	}

	err := s.store.UpdateResponse(e)
	if err != nil {
		return err
	}

	return nil
}
