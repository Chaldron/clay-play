package event

import (
	"database/sql"
	"errors"
	"fmt"
	groupPkg "github/mattfan00/jvbe/group"
	"github/mattfan00/jvbe/template"
	"log"
	"sync"
	"time"
)

type Service struct {
	store             *Store
	group             *groupPkg.Service
	eventResponseLock sync.Mutex
}

func NewService(store *Store, group *groupPkg.Service) *Service {
	return &Service{
		store: store,
		group: group,
	}
}

func eventLog(format string, s ...any) {
	log.Printf("event/event.go: %s", fmt.Sprintf(format, s...))
}

func (s *Service) GetCurrent(userId string) ([]Event, error) {
	currEvents, err := s.store.GetCurrent()
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

func (s *Service) Get(id string) (Event, error) {
	e, err := s.store.GetById(id)
	return e, err
}

func (s *Service) GetDetailed(eventId string, userId string) (EventDetailed, error) {
	event, err := s.store.GetById(eventId)
	if err != nil {
		return EventDetailed{}, err
	}

	if err = s.group.CanAccessError(event.GroupId, userId); err != nil {
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

func timeFromForm(t string, offset int) (time.Time, error) {
	r, err := time.Parse(template.FormTimeFormat, t)
	if err != nil {
		return time.Time{}, err
	}
	r = r.Add(time.Minute * time.Duration(offset))

	return r, nil
}

func (s *Service) CreateFromRequest(req CreateRequest) error {
	eventLog("CreateFromRequest req %+v", req)
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

	err = s.store.InsertOne(newEvent)
	if err != nil {
		return err
	}

	return nil
}

// TODO: handle managing the waitlist if there were people on the waitlist and capacity increased
func (s *Service) Update(req UpdateRequest) error {
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

func (s *Service) Delete(eventId string) error {
	eventLog("Delete id:%s", eventId)
	err := s.store.DeleteById(eventId)
	if err != nil {
		return err
	}

	return nil
}

var MaxAttendeeCount = 2

func (s *Service) HandleEventResponse(userId string, req RespondEventRequest) error {
	eventLog("HandleEventResponse userId:%s req:%+v", userId, req)
	s.eventResponseLock.Lock()
	defer s.eventResponseLock.Unlock()

	if req.AttendeeCount < 0 {
		return errors.New("cannot have less than 0 attendees")
	}

	if req.AttendeeCount > MaxAttendeeCount {
		return fmt.Errorf("maximum of %d plus one(s) allowed", MaxAttendeeCount-1)
	}

	e, err := s.store.GetById(req.Id)
	if err != nil {
		return err
	}

	if err = s.group.CanAccessError(e.GroupId, userId); err != nil {
		return err
	}

	existingResponse, err := s.store.GetUserResponse(req.Id, userId)
	if err != nil {
		return err
	}

	attendeeCountDelta := req.AttendeeCount
	if existingResponse != nil { // if a response exists already, need to factor the attendees in that one
		attendeeCountDelta -= existingResponse.AttendeeCount
	}

	if req.AttendeeCount == 0 { // just delete the response, I don't think it really matters to keep it in DB
		err := s.store.DeleteResponse(req.Id, userId)
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
			UserId:        userId,
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
func (s *Service) ManageWaitlist(eventId string, count int) error {
	waitlist, err := s.store.GetWaitlist(eventId, count)
	if err != nil {
		return err
	}

	err = s.store.UpdateWaitlist(waitlist)
	if err != nil {
		return err
	}

	return nil
}
