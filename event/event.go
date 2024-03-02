package event

import (
	"database/sql"
	"errors"
	"fmt"
	groupPkg "github.com/mattfan00/jvbe/group"
	"github.com/mattfan00/jvbe/template"
	"log"
	"sync"
	"time"
)

type Service interface {
	ListCurrent(string) ([]Event, error)
	Get(string) (Event, error)
	GetDetailed(string, string) (EventDetailed, error)
	Update(UpdateRequest) error
	Create(CreateRequest) error
	Delete(string) error
	HandleResponse(RespondEventRequest) error
	ManageWaitlist(string, int) error
}

type service struct {
	store             Store
	group             groupPkg.Service
	eventResponseLock sync.Mutex
}

func NewService(store Store, group groupPkg.Service) *service {
	return &service{
		store: store,
		group: group,
	}
}

func eventLog(format string, s ...any) {
	log.Printf("event/event.go: %s", fmt.Sprintf(format, s...))
}

func (s *service) ListCurrent(userId string) ([]Event, error) {
	currEvents, err := s.store.ListCurrent()
	if err != nil {
		return []Event{}, err
	}

	// filter out events you don't have access to
	filtered := []Event{}
	for _, e := range currEvents {
		ok, err := s.group.CanAccess(e.GroupId, userId)
		if err != nil {
			return []Event{}, err
		}
		if ok {
			filtered = append(filtered, e)
		}
	}

	return filtered, nil
}

func (s *service) Get(id string) (Event, error) {
	e, err := s.store.Get(id)
	return e, err
}

func (s *service) GetDetailed(eventId string, userId string) (EventDetailed, error) {
	event, err := s.store.Get(eventId)
	if err != nil {
		return EventDetailed{}, err
	}

	if err = s.group.CanAccessError(event.GroupId, userId); err != nil {
		return EventDetailed{}, err
	}

	responses, err := s.store.ListResponses(eventId)
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

func timeFromForm(t string, offset int) (time.Time, error) {
	r, err := time.Parse(template.FormTimeFormat, t)
	if err != nil {
		return time.Time{}, err
	}
	r = r.Add(time.Minute * time.Duration(offset))

	return r, nil
}

type CreateRequest struct {
	Name           string `schema:"name"`
	GroupId        string `schema:"groupId"`
	Capacity       int    `schema:"capacity"`
	Start          string `schema:"start"`
	TimezoneOffset int    `schema:"timezoneOffset"`
	Location       string `schema:"location"`
	CreatorId      string
}

func (s *service) Create(req CreateRequest) error {
	eventLog("Create req %+v", req)
	start, err := timeFromForm(req.Start, req.TimezoneOffset)
	if err != nil {
		return err
	}

	newEvent := Event{
		Name: req.Name,
		GroupId: sql.NullString{
			String: req.GroupId,
			Valid:  req.GroupId != "",
		},
		Capacity:  req.Capacity,
		Start:     start,
		Location:  req.Location,
		CreatedAt: time.Now(),
		CreatorId: req.CreatorId,
	}

	err = s.store.Create(newEvent)
	if err != nil {
		return err
	}

	return nil
}

type UpdateRequest struct {
	Id             string
	Name           string `schema:"name"`
	Capacity       int    `schema:"capacity"`
	Start          string `schema:"start"`
	TimezoneOffset int    `schema:"timezoneOffset"`
	Location       string `schema:"location"`
}

// TODO: handle managing the waitlist if there were people on the waitlist and capacity increased
func (s *service) Update(req UpdateRequest) error {
	eventLog("Update req %+v", req)
	start, err := timeFromForm(req.Start, req.TimezoneOffset)
	if err != nil {
		return err
	}

	err = s.store.Update(UpdateParams{
		Id:       req.Id,
		Name:     req.Name,
		Capacity: req.Capacity,
		Start:    start,
		Location: req.Location,
	})

	return err
}

func (s *service) Delete(eventId string) error {
	eventLog("Delete id:%s", eventId)
	err := s.store.Delete(eventId)
	if err != nil {
		return err
	}

	return nil
}

var MaxAttendeeCount = 2

type RespondEventRequest struct {
	UserId        string
	Id            string `schema:"id"`
	AttendeeCount int    `schema:"attendeeCount"`
}

func (s *service) HandleResponse(req RespondEventRequest) error {
	eventLog("HandleResponse req %+v", req)
	s.eventResponseLock.Lock()
	defer s.eventResponseLock.Unlock()

	if req.AttendeeCount < 0 {
		return errors.New("cannot have less than 0 attendees")
	}

	if req.AttendeeCount > MaxAttendeeCount {
		return fmt.Errorf("maximum of %d plus one(s) allowed", MaxAttendeeCount-1)
	}

	e, err := s.store.Get(req.Id)
	if err != nil {
		return err
	}

	if err = s.group.CanAccessError(e.GroupId, req.UserId); err != nil {
		return err
	}

	existingResponse, err := s.store.GetUserResponse(req.Id, req.UserId)
	if err != nil {
		return err
	}

	attendeeCountDelta := req.AttendeeCount
	if existingResponse != nil { // if a response exists already, need to factor the attendees in that one
		attendeeCountDelta -= existingResponse.AttendeeCount
	}

	if req.AttendeeCount == 0 { // just delete the response, I don't think it really matters to keep it in DB
		err := s.store.DeleteResponse(req.Id, req.UserId)
		if err != nil {
			return err
		}
	} else {
		// if theres no space for the response coming in, add the response to the waitlist
		// waitlist responses should ALWAYS be 1 attendee (no plus ones)
		addToWaitlist := e.SpotsLeft()-attendeeCountDelta < 0
		if addToWaitlist && req.AttendeeCount > 1 {
			return errors.New("no plus ones when adding to waitlist")
		}

		er := EventResponse{
			EventId:       req.Id,
			UserId:        req.UserId,
			AttendeeCount: req.AttendeeCount,
			OnWaitlist:    addToWaitlist,
		}

		err = s.store.UpdateResponse(er)
		if err != nil {
			return err
		}
	}

	fromAttendee := !(existingResponse != nil && existingResponse.OnWaitlist)
	eventLog("spots left:%d delta:%d attendee:%t", e.SpotsLeft(), attendeeCountDelta, fromAttendee)
	// manage waitlist ugh
	// only need to manage it if the event had no spots left and spots freed up from main attendee list
	if e.SpotsLeft() == 0 && attendeeCountDelta < 0 && fromAttendee {
		err := s.ManageWaitlist(req.Id, attendeeCountDelta*-1)
		if err != nil {
			return err
		}
	}

	return nil
}

// take people off the waitlist based off count
func (s *service) ManageWaitlist(eventId string, count int) error {
	waitlist, err := s.store.ListWaitlist(eventId, count)
	if err != nil {
		return err
	}

	err = s.store.UpdateWaitlist(waitlist)
	if err != nil {
		return err
	}

	return nil
}
