package user

import (
	"errors"
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) HandleFromExternal(externalUser ExternalUser) (User, error) {
	user, err := s.store.GetByExternalId(externalUser.Id)
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
