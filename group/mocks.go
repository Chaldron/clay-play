package group

import (
	"database/sql"
	"github.com/mattfan00/jvbe/user"

	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) Create(r CreateRequest) (string, error) {
	args := m.Called(r)
	return args.String(0), args.Error(1)
}

func (m *MockService) Get(s string) (Group, error) {
	args := m.Called(s)
	return args.Get(0).(Group), args.Error(1)
}

func (m *MockService) GetDetailed(s string, u user.SessionUser) (GroupDetailed, error) {
	args := m.Called(s, u)
	return args.Get(0).(GroupDetailed), args.Error(1)
}

func (m *MockService) List() ([]Group, error) {
	args := m.Called()
	return args.Get(0).([]Group), args.Error(1)
}

func (m *MockService) AddMember(s1 string, s2 string) error {
	args := m.Called(s1, s2)
	return args.Error(0)
}

func (m *MockService) AddMemberFromInvite(s1 string, s2 string) (Group, error) {
	args := m.Called(s1, s2)
	return args.Get(0).(Group), args.Error(1)
}

func (m *MockService) Update(r UpdateRequest) error {
	args := m.Called(r)
	return args.Error(9)
}

func (m *MockService) Delete(s string) error {
	args := m.Called(s)
	return args.Error(0)
}

func (m *MockService) RemoveMember(s1 string, s2 string) error {
	args := m.Called(s1, s2)
	return args.Error(0)
}

func (m *MockService) CreateAndAddMember(r CreateRequest) error {
	args := m.Called(r)
	return args.Error(0)
}

func (m *MockService) CanAccess(s1 sql.NullString, s2 string) (bool, error) {
	args := m.Called(s1, s2)
	return args.Bool(0), args.Error(1)
}

func (m *MockService) CanAccessError(s1 sql.NullString, s2 string) error {
	args := m.Called(s1, s2)
	return args.Error(0)
}

func (m *MockService) RefreshInviteId(s string) error {
	args := m.Called(s)
	return args.Error(0)
}
