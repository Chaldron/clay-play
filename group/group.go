package group

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattfan00/jvbe/config"
	"github.com/mattfan00/jvbe/user"
	"log"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

var (
	ErrNoAccess = errors.New("you do not have access")
)

type Service interface {
	Create(CreateRequest) (string, error)
	Get(string) (Group, error)
	GetDetailed(string, user.SessionUser) (GroupDetailed, error)
	List() ([]Group, error)
	AddMember(string, string) error
	AddMemberFromInvite(string, string) (Group, error)
	Update(UpdateRequest) error
	Delete(string) error
	RemoveMember(string, string) error
	CreateAndAddMember(CreateRequest) error
	CanAccess(sql.NullString, string) (bool, error)
	CanAccessError(sql.NullString, string) error
	RefreshInviteId(string) error
}

type service struct {
	store Store
	conf  *config.Config
}

func NewService(store Store, conf *config.Config) *service {
	return &service{
		store: store,
		conf:  conf,
	}
}

func groupLog(format string, s ...any) {
	log.Printf("group/group.go: %s", fmt.Sprintf(format, s...))
}

type CreateRequest struct {
	CreatorId string
	Name      string `schema:"name"`
}

func (s *service) Create(req CreateRequest) (string, error) {
	groupLog("Create req %+v", req)
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

func (s *service) Get(id string) (Group, error) {
	g, err := s.store.Get(id)
	return g, err
}

func (s *service) GetDetailed(id string, user user.SessionUser) (GroupDetailed, error) {
	if !user.CanModifyGroup() {
		if err := s.CanAccessError(sql.NullString{
			String: id,
			Valid:  true,
		}, user.Id); err != nil {
			return GroupDetailed{}, err
		}
	}

	g, err := s.store.Get(id)
	if err != nil {
		return GroupDetailed{}, err
	}

	m, err := s.store.ListMembers(id)
	if err != nil {
		return GroupDetailed{}, err
	}

	return GroupDetailed{
		Group:   g,
		Members: m,
	}, err
}

func (s *service) List() ([]Group, error) {
	g, err := s.store.List()
	return g, err
}

func (s *service) AddMember(groupId string, userId string) error {
	groupLog("AddMember groupId:%s userId:%s", groupId, userId)
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

func (s *service) AddMemberFromInvite(inviteId string, userId string) (Group, error) {
	groupLog("AddMemberFromInvite inviteId:%s userId:%s", inviteId, userId)
	g, err := s.store.GetByInvite(inviteId)
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

func (s *service) Update(req UpdateRequest) error {
	groupLog("Update req %+v", req)
	err := s.store.Update(UpdateParams{
		Id:   req.Id,
		Name: req.Name,
	})
	return err
}

func (s *service) Delete(id string) error {
	groupLog("Delete id:%s", id)
	err := s.store.Delete(id)
	return err
}

func (s *service) RemoveMember(groupId string, userId string) error {
	groupLog("RemoveMember groupId:%s userId:%s", groupId, userId)
	g, err := s.store.Get(groupId)
	if g.CreatorId == userId {
		return errors.New("cannot remove the creator of a group")
	}

	err = s.store.RemoveMember(groupId, userId)
	return err
}

func (s *service) CreateAndAddMember(req CreateRequest) error {
	groupLog("CreateAndAddMember req %+v", req)
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

func (s *service) CanAccess(groupId sql.NullString, userId string) (bool, error) {
	if !groupId.Valid { // if NULL, then public group
		return true, nil
	}

	exists, err := s.store.HasMember(groupId.String, userId)
	return exists, err
}

func (s *service) CanAccessError(groupId sql.NullString, userId string) error {
	ok, err := s.CanAccess(groupId, userId)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNoAccess
	}

	return nil
}

func (s *service) RefreshInviteId(groupId string) error {
	inviteId, err := gonanoid.New()
	if err != nil {
		return err
	}

	return s.store.RefreshInviteId(groupId, inviteId)
}
