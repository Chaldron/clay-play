package event_test

import (
	"testing"
	"time"

	"github.com/mattfan00/jvbe/db"
	"github.com/mattfan00/jvbe/event"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestHandleResponse(t *testing.T) {
	t.Run("NegativeAttendeesError", func(t *testing.T) {
		err := event.NewService(nil).HandleResponse(event.HandleResponseParams{
			AttendeeCount: -1,
		})

		assert.Error(t, err)
	})
}

func TestList(t *testing.T) {
	day := 24 * time.Hour

	t.Run("Ok", func(t *testing.T) {
		db := db.TestingConnect(t)
		defer db.Close()
		eventService := event.NewService(db)

		MustCreate(t, db, event.CreateParams{Name: "one", Start: time.Now()})
		MustCreate(t, db, event.CreateParams{Name: "two", Start: time.Now().Add(day)})

		events, err := eventService.List(event.ListFilter{})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(events))
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
		assert.Equal(t, 1, len(events))
		assert.Equal(t, "upcoming", events[0].Name)
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
		assert.Equal(t, 1, len(events))
		assert.Equal(t, "past", events[0].Name)
	})
}

func MustCreate(t testing.TB, db *db.DB, p event.CreateParams) {
	t.Helper()
	err := event.NewService(db).Create(p)
	if err != nil {
		t.Fatal(err)
	}
}
