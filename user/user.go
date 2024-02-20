package user

import (
	"errors"
	"fmt"
	"log"
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{
		store: store,
	}
}

func userLog(format string, s ...any) {
	log.Printf("user/user.go: %s", fmt.Sprintf(format, s...))
}

func (s *Service) HandleFromExternal(externalUser ExternalUser) (User, error) {
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
