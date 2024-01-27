package user

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) HandleFromExternal(externalUser ExternalUser) (string, error) {
	id, err := s.store.GetIdByExternalId(externalUser.Id)
	if err != nil {
		return "", err
	}

	// if id is blank then need to create
	if id == "" {
		newId, err := s.store.CreateFromExternal(externalUser)
		if err != nil {
			return "", err
		}

		return newId, nil
	} else {
		return id, nil
	}
}
