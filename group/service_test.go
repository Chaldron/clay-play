package group_test

import (
	"testing"

	"github.com/Chaldron/clay-play/db"
	"github.com/Chaldron/clay-play/event"
	"github.com/Chaldron/clay-play/group"
	"github.com/Chaldron/clay-play/user"
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

	_, err = eventService.Create(event.CreateParams{GroupId: groupId})
	if err != nil {
		t.Fatal(t)
	}
	events, err := eventService.List(event.ListFilter{})
	if err != nil {
		t.Fatal(t)
	}

	// at this point, u1 is part of the group and u2 is not
	// so u1 should be able to see the event
	u1Events, err := groupService.FilterEventsUserCanAccess(events.Events, u1.Id)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(u1Events))

	// and u2 should not
	u2Events, err := groupService.FilterEventsUserCanAccess(events.Events, u2.Id)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(u2Events))
}
