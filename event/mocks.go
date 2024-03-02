package event

import (
    "github.com/stretchr/testify/mock"
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) Create(e Event) error {
	args := m.Called(e)
	return args.Error(0)
}

func (m *MockStore) Update(p UpdateParams) error {
	args := m.Called(p)
	return args.Error(0)
}

func (m *MockStore) ListCurrent() ([]Event, error) {
	args := m.Called()
	return args.Get(0).([]Event), args.Error(1)
}

func (m *MockStore) Get(s string) (Event, error) {
	args := m.Called(s)
	return args.Get(0).(Event), args.Error(1)
}

func (m *MockStore) UpdateResponse(r EventResponse) error {
	args := m.Called(r)
	return args.Error(0)
}

func (m *MockStore) DeleteResponse(s1 string, s2 string) error {
	args := m.Called(s1, s2)
	return args.Error(0)
}

func (m *MockStore) ListResponses(s string) ([]EventResponse, error) {
	args := m.Called(s)
	return args.Get(0).([]EventResponse), args.Error(1)
}

func (m *MockStore) ListWaitlist(s string, i int) ([]EventResponse, error) {
	args := m.Called(s, i)
	return args.Get(0).([]EventResponse), args.Error(1)
}

func (m *MockStore) UpdateWaitlist(r []EventResponse) error {
	args := m.Called(r)
	return args.Error(0)
}

func (m *MockStore) GetUserResponse(s1 string, s2 string) (*EventResponse, error) {
	args := m.Called(s1, s2)
	return args.Get(0).(*EventResponse), args.Error(1)
}

func (m *MockStore) Delete(s string) error {
	args := m.Called(s)
	return args.Error(0)
}
