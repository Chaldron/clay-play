package tests

import (
	"database/sql"

	"github.com/mattfan00/jvbe/event"
	"github.com/mattfan00/jvbe/group"
	"github.com/mattfan00/jvbe/user"

	"github.com/stretchr/testify/mock"
)

type MockEventStore struct {
	mock.Mock
}

func (m *MockEventStore) Create(e event.Event) error {
	args := m.Called(e)
	return args.Error(0)
}

func (m *MockEventStore) Update(p event.UpdateParams) error {
	args := m.Called(p)
	return args.Error(0)
}

func (m *MockEventStore) ListCurrent() ([]event.Event, error) {
	args := m.Called()
	return args.Get(0).([]event.Event), args.Error(1)
}

func (m *MockEventStore) Get(s string) (event.Event, error) {
	args := m.Called(s)
	return args.Get(0).(event.Event), args.Error(1)
}

func (m *MockEventStore) UpdateResponse(r event.EventResponse) error {
	args := m.Called(r)
	return args.Error(0)
}

func (m *MockEventStore) DeleteResponse(s1 string, s2 string) error {
	args := m.Called(s1, s2)
	return args.Error(0)
}

func (m *MockEventStore) ListResponses(s string) ([]event.EventResponse, error) {
	args := m.Called(s)
	return args.Get(0).([]event.EventResponse), args.Error(1)
}

func (m *MockEventStore) ListWaitlist(s string, i int) ([]event.EventResponse, error) {
	args := m.Called(s, i)
	return args.Get(0).([]event.EventResponse), args.Error(1)
}

func (m *MockEventStore) UpdateWaitlist(r []event.EventResponse) error {
	args := m.Called(r)
	return args.Error(0)
}

func (m *MockEventStore) GetUserResponse(s1 string, s2 string) (*event.EventResponse, error) {
	args := m.Called(s1, s2)
	return args.Get(0).(*event.EventResponse), args.Error(1)
}

func (m *MockEventStore) Delete(s string) error {
	args := m.Called(s)
	return args.Error(0)
}

type MockGroupService struct {
	mock.Mock
}

func (m *MockGroupService) Create(r group.CreateRequest) (string, error) {
	args := m.Called(r)
	return args.String(0), args.Error(1)
}

func (m *MockGroupService) Get(s string) (group.Group, error) {
	args := m.Called(s)
	return args.Get(0).(group.Group), args.Error(1)
}

func (m *MockGroupService) GetDetailed(s string, u user.SessionUser) (group.GroupDetailed, error) {
	args := m.Called(s, u)
	return args.Get(0).(group.GroupDetailed), args.Error(1)
}

func (m *MockGroupService) List() ([]group.Group, error) {
	args := m.Called()
	return args.Get(0).([]group.Group), args.Error(1)
}

func (m *MockGroupService) AddMember(s1 string, s2 string) error {
	args := m.Called(s1, s2)
	return args.Error(0)
}

func (m *MockGroupService) AddMemberFromInvite(s1 string, s2 string) (group.Group, error) {
	args := m.Called(s1, s2)
	return args.Get(0).(group.Group), args.Error(1)
}

func (m *MockGroupService) Update(r group.UpdateRequest) error {
	args := m.Called(r)
	return args.Error(9)
}

func (m *MockGroupService) Delete(s string) error {
	args := m.Called(s)
	return args.Error(0)
}

func (m *MockGroupService) RemoveMember(s1 string, s2 string) error {
	args := m.Called(s1, s2)
	return args.Error(0)
}

func (m *MockGroupService) CreateAndAddMember(r group.CreateRequest) error {
	args := m.Called(r)
	return args.Error(0)
}

func (m *MockGroupService) CanAccess(s1 sql.NullString, s2 string) (bool, error) {
	args := m.Called(s1, s2)
	return args.Bool(0), args.Error(1)
}

func (m *MockGroupService) CanAccessError(s1 sql.NullString, s2 string) error {
	args := m.Called(s1, s2)
	return args.Error(0)
}

func (m *MockGroupService) RefreshInviteId(s string) error {
	args := m.Called(s)
	return args.Error(0)
}
