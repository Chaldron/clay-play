package user

import (
	"errors"
	"fmt"
	"log"
)

type Service interface {
	HandleFromExternal(ExternalUser) (User, error)
}

type service struct {
	store Store
}

func NewService(store Store) *service {
	return &service{
		store: store,
	}
}

func userLog(format string, s ...any) {
	log.Printf("user/user.go: %s", fmt.Sprintf(format, s...))
}

func (s *service) HandleFromExternal(externalUser ExternalUser) (User, error) {
	userLog("HandleFromExternal externalUser %+v", externalUser)
	user, err := s.store.GetByExternal(externalUser.Id)
	// if cant retrieve user, then need to create
	if errors.Is(err, ErrNoUser) {
		user, err = s.store.CreateFromExternal(externalUser)
		if err != nil {
			return User{}, err
		}
	} else if err != nil {
		return User{}, err
	}

	return user, nil
}
