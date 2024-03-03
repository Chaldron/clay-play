package event_test

import (
	"testing"

	"github.com/mattfan00/jvbe/event"
	"github.com/mattfan00/jvbe/tests"

	"github.com/stretchr/testify/assert"
)

func TestHandleResponseNegativeAttendees(t *testing.T) {
	eventStore := new(tests.MockEventStore)
	groupService := new(tests.MockGroupService)
	s := event.NewService(eventStore, groupService)

	r := event.RespondEventRequest{
		UserId:        "1",
		Id:            "a",
		AttendeeCount: -1,
	}
	err := s.HandleResponse(r)

	assert.Error(t, err)
}
