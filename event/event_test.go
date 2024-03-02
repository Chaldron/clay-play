package event

import (
	"github.com/mattfan00/jvbe/group"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleResponseNegativeAttendees(t *testing.T) {
	eventStore := new(MockStore)
	groupService := new(group.MockService)
	s := NewService(eventStore, groupService)

	r := RespondEventRequest{
		UserId:        "1",
		Id:            "a",
		AttendeeCount: -1,
	}
	err := s.HandleResponse(r)

	assert.Error(t, err)
}
