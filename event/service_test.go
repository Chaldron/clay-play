package event_test

import (
	"testing"
	"time"

	"github.com/mattfan00/jvbe/db"
	"github.com/mattfan00/jvbe/event"
	"github.com/mattfan00/jvbe/group"
	"github.com/mattfan00/jvbe/user"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

var day = 24 * time.Hour

func TestHandleResponse(t *testing.T) {
	t.Run("NegativeAttendeesError", func(t *testing.T) {
		err := event.NewService(nil).HandleResponse(event.HandleResponseParams{
			AttendeeCount: -1,
		})

		assert.Error(t, err)
	})

	t.Run("IsPastError", func(t *testing.T) {
		db := db.TestingConnect(t)
		defer db.Close()
		eventService := event.NewService(db)
		userService := user.NewService(db)

		u, err := userService.Create(user.CreateParams{})
		if err != nil {
			t.Fatal(err)
		}

		id := MustCreate(t, db, event.CreateParams{CreatorId: u.Id, Start: time.Now().Add(-day)})

		err = eventService.HandleResponse(event.HandleResponseParams{
			UserId:        u.Id,
			Id:            id,
			AttendeeCount: 0,
		})
		assert.Error(t, err)
	})

	t.Run("ManageWaitlist", func(t *testing.T) {
		db := db.TestingConnect(t)
		defer db.Close()
		eventService := event.NewService(db)
		userService := user.NewService(db)

		u1, err := userService.Create(user.CreateParams{})
		if err != nil {
			t.Fatal(err)
		}
		u2, err := userService.Create(user.CreateParams{})
		if err != nil {
			t.Fatal(err)
		}
		u3, err := userService.Create(user.CreateParams{})
		if err != nil {
			t.Fatal(err)
		}
		id := MustCreate(t, db, event.CreateParams{
			CreatorId: u1.Id,
			Start:     time.Now().Add(day),
			Capacity:  2,
		})

		err = eventService.HandleResponse(event.HandleResponseParams{
			UserId:        u1.Id,
			Id:            id,
			AttendeeCount: 1,
		})
		if err != nil {
			t.Fatal(err)
		}
		err = eventService.HandleResponse(event.HandleResponseParams{
			UserId:        u2.Id,
			Id:            id,
			AttendeeCount: 1,
		})
		if err != nil {
			t.Fatal(err)
		}
		err = eventService.HandleResponse(event.HandleResponseParams{
			UserId:        u3.Id,
			Id:            id,
			AttendeeCount: 1,
		})
		if err != nil {
			t.Fatal(err)
		}

		responses, err := eventService.ListResponses(id)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 3, len(responses))

		assert.Equal(t, false, responses[0].OnWaitlist)
		assert.Equal(t, u1.Id, responses[0].UserId)

		assert.Equal(t, false, responses[1].OnWaitlist)
		assert.Equal(t, u2.Id, responses[1].UserId)

		assert.Equal(t, true, responses[2].OnWaitlist)
		assert.Equal(t, u3.Id, responses[2].UserId)

		err = eventService.HandleResponse(event.HandleResponseParams{
			UserId:        u2.Id,
			Id:            id,
			AttendeeCount: 0,
		})
		if err != nil {
			t.Fatal(err)
		}

		responses, err = eventService.ListResponses(id)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 2, len(responses))

		assert.Equal(t, false, responses[0].OnWaitlist)
		assert.Equal(t, u1.Id, responses[0].UserId)

		assert.Equal(t, false, responses[1].OnWaitlist)
		assert.Equal(t, u3.Id, responses[1].UserId)
	})

	t.Run("OnWaitlistEvenIfSpotsLeft", func(t *testing.T) {
		db := db.TestingConnect(t)
		defer db.Close()
		eventService := event.NewService(db)
		userService := user.NewService(db)

		u1, err := userService.Create(user.CreateParams{})
		if err != nil {
			t.Fatal(err)
		}
		u2, err := userService.Create(user.CreateParams{})
		if err != nil {
			t.Fatal(err)
		}
		u3, err := userService.Create(user.CreateParams{})
		if err != nil {
			t.Fatal(err)
		}
		id := MustCreate(t, db, event.CreateParams{
			CreatorId: u1.Id,
			Start:     time.Now().Add(day),
			Capacity:  2,
		})

		err = eventService.HandleResponse(event.HandleResponseParams{
			UserId:        u1.Id,
			Id:            id,
			AttendeeCount: 1,
		})
		if err != nil {
			t.Fatal(err)
		}
		err = eventService.HandleResponse(event.HandleResponseParams{
			UserId:        u2.Id,
			Id:            id,
			AttendeeCount: 2,
		})
		if err != nil {
			t.Fatal(err)
		}

		responses, err := eventService.ListResponses(id)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 2, len(responses))

		assert.Equal(t, false, responses[0].OnWaitlist)
		assert.Equal(t, u1.Id, responses[0].UserId)

		assert.Equal(t, true, responses[1].OnWaitlist)
		assert.Equal(t, u2.Id, responses[1].UserId)

		err = eventService.HandleResponse(event.HandleResponseParams{
			UserId:        u3.Id,
			Id:            id,
			AttendeeCount: 1,
		})
		if err != nil {
			t.Fatal(err)
		}

		responses, err = eventService.ListResponses(id)
		if err != nil {
			t.Fatal(err)
		}

		e, err := eventService.Get(id)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 3, len(responses))
		assert.Equal(t, 1, e.SpotsLeft())

		assert.Equal(t, false, responses[0].OnWaitlist)
		assert.Equal(t, u1.Id, responses[0].UserId)

		assert.Equal(t, true, responses[1].OnWaitlist)
		assert.Equal(t, u2.Id, responses[1].UserId)

		assert.Equal(t, true, responses[2].OnWaitlist)
		assert.Equal(t, u3.Id, responses[2].UserId)
	})
}

func TestGet(t *testing.T) {
	t.Run("IsPast", func(t *testing.T) {
		db := db.TestingConnect(t)
		defer db.Close()
		eventService := event.NewService(db)
		userService := user.NewService(db)

		u, err := userService.Create(user.CreateParams{})
		if err != nil {
			t.Fatal(err)
		}

		id := MustCreate(t, db, event.CreateParams{CreatorId: u.Id, Start: time.Now().Add(-day)})

		e, err := eventService.Get(id)
		assert.NoError(t, err)
		assert.Equal(t, true, e.IsPast)

		id = MustCreate(t, db, event.CreateParams{CreatorId: u.Id, Start: time.Now().Add(day)})

		event, err := eventService.Get(id)
		assert.NoError(t, err)
		assert.Equal(t, false, event.IsPast)
	})
}

func TestList(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		db := db.TestingConnect(t)
		defer db.Close()
		eventService := event.NewService(db)

		MustCreate(t, db, event.CreateParams{Name: "one", Start: time.Now()})
		MustCreate(t, db, event.CreateParams{Name: "two", GroupId: "1", Start: time.Now().Add(day)})

		events, err := eventService.List(event.ListFilter{})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(events.Events))
	})

	t.Run("FilterUpcoming", func(t *testing.T) {
		db := db.TestingConnect(t)
		defer db.Close()
		eventService := event.NewService(db)

		now := time.Now()
		MustCreate(t, db, event.CreateParams{Name: "past", Start: now.Add(-day)})
		MustCreate(t, db, event.CreateParams{Name: "upcoming", Start: now.Add(day)})

		events, err := eventService.List(event.ListFilter{Upcoming: true})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(events.Events))
		assert.Equal(t, "upcoming", events.Events[0].Name)
	})

	t.Run("FilterPast", func(t *testing.T) {
		db := db.TestingConnect(t)
		defer db.Close()
		eventService := event.NewService(db)

		now := time.Now()
		MustCreate(t, db, event.CreateParams{Name: "past", Start: now.Add(-day)})
		MustCreate(t, db, event.CreateParams{Name: "upcoming", Start: now.Add(day)})

		events, err := eventService.List(event.ListFilter{Past: true})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(events.Events))
		assert.Equal(t, "past", events.Events[0].Name)
	})

	t.Run("FilterUserIdCanAccess", func(t *testing.T) {
		db := db.TestingConnect(t)
		defer db.Close()
		eventService := event.NewService(db)
		groupService := group.NewService(db)
		userService := user.NewService(db)

		u, err := userService.Create(user.CreateParams{})
		if err != nil {
			t.Fatal(err)
		}

		g1, err := groupService.CreateAndAddMember(group.CreateParams{CreatorId: u.Id})
		if err != nil {
			t.Fatal(err)
		}
		g2, err := groupService.CreateAndAddMember(group.CreateParams{CreatorId: u.Id})
		if err != nil {
			t.Fatal(err)
		}

		MustCreate(t, db, event.CreateParams{GroupId: ""})
		MustCreate(t, db, event.CreateParams{GroupId: g1})
		MustCreate(t, db, event.CreateParams{GroupId: g2})
		// cannot include this one since user does not belong to this group
		MustCreate(t, db, event.CreateParams{GroupId: "1"})

		events, err := eventService.List(event.ListFilter{UserId: u.Id})
		assert.NoError(t, err)
		assert.Equal(t, 3, len(events.Events))
		for _, e := range events.Events {
			assert.NotEqual(t, "1", e.GroupId)
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Run("ManageWaitlist", func(t *testing.T) {
		t.Run("ReduceCapacity", func(t *testing.T) {
			db := db.TestingConnect(t)
			defer db.Close()
			eventService := event.NewService(db)
			userService := user.NewService(db)

			u1, err := userService.Create(user.CreateParams{})
			if err != nil {
				t.Fatal(err)
			}
			u2, err := userService.Create(user.CreateParams{})
			if err != nil {
				t.Fatal(err)
			}
			eventId := MustCreate(t, db, event.CreateParams{
				Start:     time.Now().Add(24 * time.Hour),
				CreatorId: u1.Id,
				Capacity:  2,
			})

			MustHandleResponse(t, db, event.HandleResponseParams{
				UserId:        u1.Id,
				Id:            eventId,
				AttendeeCount: 1,
			})
			MustHandleResponse(t, db, event.HandleResponseParams{
				UserId:        u2.Id,
				Id:            eventId,
				AttendeeCount: 1,
			})

			responses, err := eventService.ListResponses(eventId)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, 2, len(responses))
			assert.Equal(t, false, responses[0].OnWaitlist)
			assert.Equal(t, false, responses[1].OnWaitlist)

			err = eventService.Update(event.UpdateParams{
				Id:       eventId,
				Capacity: 1,
			})
			if err != nil {
				t.Fatal(err)
			}

			responses, err = eventService.ListResponses(eventId)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, 2, len(responses))
			assert.Equal(t, false, responses[0].OnWaitlist)
			assert.Equal(t, true, responses[1].OnWaitlist)
		})

		t.Run("ReduceCapacityUneven", func(t *testing.T) {
			db := db.TestingConnect(t)
			defer db.Close()
			eventService := event.NewService(db)
			userService := user.NewService(db)

			u1, err := userService.Create(user.CreateParams{})
			if err != nil {
				t.Fatal(err)
			}
			u2, err := userService.Create(user.CreateParams{})
			if err != nil {
				t.Fatal(err)
			}
			eventId := MustCreate(t, db, event.CreateParams{
				Start:     time.Now().Add(24 * time.Hour),
				CreatorId: u1.Id,
				Capacity:  3,
			})

			MustHandleResponse(t, db, event.HandleResponseParams{
				UserId:        u1.Id,
				Id:            eventId,
				AttendeeCount: 1,
			})
			MustHandleResponse(t, db, event.HandleResponseParams{
				UserId:        u2.Id,
				Id:            eventId,
				AttendeeCount: 2,
			})

			responses, err := eventService.ListResponses(eventId)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, 2, len(responses))
			assert.Equal(t, false, responses[0].OnWaitlist)
			assert.Equal(t, false, responses[1].OnWaitlist)

			err = eventService.Update(event.UpdateParams{
				Id:       eventId,
				Capacity: 2,
			})
			if err != nil {
				t.Fatal(err)
			}

			responses, err = eventService.ListResponses(eventId)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, 2, len(responses))
			assert.Equal(t, false, responses[0].OnWaitlist)
			assert.Equal(t, true, responses[1].OnWaitlist)
		})

		t.Run("IncreaseCapacity", func(t *testing.T) {
			db := db.TestingConnect(t)
			defer db.Close()
			eventService := event.NewService(db)
			userService := user.NewService(db)

			u1, err := userService.Create(user.CreateParams{})
			if err != nil {
				t.Fatal(err)
			}
			u2, err := userService.Create(user.CreateParams{})
			if err != nil {
				t.Fatal(err)
			}
			eventId := MustCreate(t, db, event.CreateParams{
				Start:     time.Now().Add(24 * time.Hour),
				CreatorId: u1.Id,
				Capacity:  1,
			})

			MustHandleResponse(t, db, event.HandleResponseParams{
				UserId:        u1.Id,
				Id:            eventId,
				AttendeeCount: 1,
			})
			MustHandleResponse(t, db, event.HandleResponseParams{
				UserId:        u2.Id,
				Id:            eventId,
				AttendeeCount: 1,
			})

			responses, err := eventService.ListResponses(eventId)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, 2, len(responses))
			assert.Equal(t, false, responses[0].OnWaitlist)
			assert.Equal(t, true, responses[1].OnWaitlist)

			err = eventService.Update(event.UpdateParams{
				Id:       eventId,
				Capacity: 2,
			})
			if err != nil {
				t.Fatal(err)
			}

			responses, err = eventService.ListResponses(eventId)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, 2, len(responses))
			assert.Equal(t, false, responses[0].OnWaitlist)
			assert.Equal(t, false, responses[1].OnWaitlist)
		})
	})
}

func MustCreate(t testing.TB, db *db.DB, p event.CreateParams) string {
	t.Helper()
	id, err := event.NewService(db).Create(p)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func MustHandleResponse(t testing.TB, db *db.DB, p event.HandleResponseParams) {
	t.Helper()
	err := event.NewService(db).HandleResponse(p)
	if err != nil {
		t.Fatal(err)
	}
}
