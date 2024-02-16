package group

import (
	"database/sql"
	"errors"
	"github/mattfan00/jvbe/config"
	"github/mattfan00/jvbe/user"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

var (
	ErrNoAccess = errors.New("you do not have access")
)

type Service struct {
	store *Store
	conf  *config.Config
}

func NewService(store *Store, conf *config.Config) *Service {
	return &Service{
		store: store,
		conf:  conf,
	}
}

type CreateRequest struct {
	CreatorId string
	Name      string `schema:"name"`
}

func (s *Service) Create(req CreateRequest) (string, error) {
	id, err := gonanoid.New()
	if err != nil {
		return "", err
	}

	inviteId, err := gonanoid.New()
	if err != nil {
		return "", err
	}

	err = s.store.Create(CreateParams{
		Id:        id,
		CreatorId: req.CreatorId,
		Name:      req.Name,
		InviteId:  inviteId,
	})
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *Service) Get(id string) (Group, error) {
	g, err := s.store.Get(id)
	return g, err
}

func (s *Service) GetDetailed(id string, user user.SessionUser) (GroupDetailed, error) {
    if !user.CanModifyGroup() {
        if err := s.CanAccessError(sql.NullString{
            String: id,
            Valid: true,
        }, user.Id); err != nil {
            return GroupDetailed{}, err
        }
    }

	g, err := s.store.Get(id)
	if err != nil {
		return GroupDetailed{}, err
	}

	m, err := s.store.GetMembers(id)
	if err != nil {
		return GroupDetailed{}, err
	}

	return GroupDetailed{
		Group:   g,
		Members: m,
	}, err
}

func (s *Service) List() ([]Group, error) {
	g, err := s.store.List()
	return g, err
}

func (s *Service) AddMember(groupId string, userId string) error {
	exists, err := s.store.HasMember(groupId, userId)
	if err != nil {
		return err
	}
	if !exists {
		err = s.store.AddMember(groupId, userId)
		if err != nil {
			return err
		}
	}

	return err
}

func (s *Service) AddMemberFromInvite(inviteId string, userId string) (Group, error) {
	g, err := s.store.GetByInviteId(inviteId)
	if err != nil {
		return Group{}, err
	}

	err = s.AddMember(g.Id, userId)
	if err != nil {
		return Group{}, err
	}

	return g, nil
}

type UpdateRequest struct {
	Id   string
	Name string `schema:"name"`
}

func (s *Service) Update(req UpdateRequest) error {
	err := s.store.Update(UpdateParams{
		Id:   req.Id,
		Name: req.Name,
	})
	return err
}

func (s *Service) Delete(id string) error {
	err := s.store.Delete(id)
	return err
}

func (s *Service) RemoveMember(groupId string, userId string) error {
	g, err := s.store.Get(groupId)
	if g.CreatorId == userId {
		return errors.New("cannot remove the creator of a group")
	}

	err = s.store.RemoveMember(groupId, userId)
	return err
}

func (s *Service) CreateAndAddMember(req CreateRequest) error {
	groupId, err := s.Create(req)
	if err != nil {
		return err
	}

	err = s.AddMember(groupId, req.CreatorId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) CanAccess(groupId sql.NullString, userId string) (bool, error) {
	if !groupId.Valid { // if NULL, then public group
		return true, nil
	}

	exists, err := s.store.HasMember(groupId.String, userId)
	return exists, err
}

func (s *Service) CanAccessError(groupId sql.NullString, userId string) error {
	ok, err := s.CanAccess(groupId, userId)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNoAccess
	}

	return nil
}
