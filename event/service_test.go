package event_test

import (
	"testing"

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
