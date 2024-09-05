package auditlog_test

import (
	"testing"

	"github.com/Chaldron/clay-play/auditlog"
	"github.com/Chaldron/clay-play/db"
	"github.com/Chaldron/clay-play/user"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	db := db.TestingConnect(t)
	defer db.Close()
	auditlogService := auditlog.NewService(db)
	userService := user.NewService(db)

	u, err := userService.Create(user.CreateParams{FullName: "name"})
	if err != nil {
		t.Fatal(err)
	}

	err = auditlogService.Create(u.Id, "test")
	if err != nil {
		t.Fatal(err)
	}

	al, _, err := auditlogService.List(auditlog.ListFilter{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(al))
	assert.Equal(t, u.Id, al[0].UserId)
	assert.Equal(t, "name", al[0].UserFullName)
	assert.Equal(t, "test", al[0].Description)
}

func TestList(t *testing.T) {
	db := db.TestingConnect(t)
	defer db.Close()
	auditlogService := auditlog.NewService(db)
	userService := user.NewService(db)

	al, count, err := auditlogService.List(auditlog.ListFilter{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 0, count)

	u1, err := userService.Create(user.CreateParams{FullName: "one"})
	if err != nil {
		t.Fatal(err)
	}
	u2, err := userService.Create(user.CreateParams{FullName: "two"})
	if err != nil {
		t.Fatal(err)
	}

	err = auditlogService.Create(u1.Id, "test1")
	if err != nil {
		t.Fatal(err)
	}
	err = auditlogService.Create(u2.Id, "test2")
	if err != nil {
		t.Fatal(err)
	}

	al, count, err = auditlogService.List(auditlog.ListFilter{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, len(al))
	assert.Equal(t, 2, count)
	assert.Equal(t, u2.Id, al[0].UserId)
	assert.Equal(t, u1.Id, al[1].UserId)
}
