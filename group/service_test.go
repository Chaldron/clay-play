package group_test

import (
	"testing"

	"github.com/mattfan00/jvbe/db"
	"github.com/mattfan00/jvbe/event"
	"github.com/mattfan00/jvbe/group"
	"github.com/mattfan00/jvbe/user"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestFilterEventsUserCanAccess(t *testing.T) {
	db := db.TestingConnect(t)
	defer db.Close()

	groupService := group.NewService(db)
	eventService := event.NewService(db)
	userService := user.NewService(db)

	u1, err := userService.Create(user.CreateParams{})
	if err != nil {
		t.Fatal(t)
	}
	u2, err := userService.Create(user.CreateParams{})
	if err != nil {
		t.Fatal(t)
	}

	groupId, err := groupService.CreateAndAddMember(group.CreateParams{
		CreatorId: u1.Id,
	})
	if err != nil {
		t.Fatal(t)
	}

	err = eventService.Create(event.CreateParams{GroupId: groupId})
	if err != nil {
		t.Fatal(t)
	}
	events, err := eventService.List(event.ListFilter{})
	if err != nil {
		t.Fatal(t)
	}

	// at this point, u1 is part of the group and u2 is not
	// so u1 should be able to see the event
	u1Events, err := groupService.FilterEventsUserCanAccess(events, u1.Id)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(u1Events))

	// and u2 should not
	u2Events, err := groupService.FilterEventsUserCanAccess(events, u2.Id)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(u2Events))
}
