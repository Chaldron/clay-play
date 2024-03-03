package user

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type Service interface {
	Get(string) (User, error)
	HandleFromExternal(ExternalUser) (User, error)
	GetReview(string) (UserReview, error)
	UpdateReview(UpdateReviewRequest) error
	ListReviews() ([]UserReview, error)
	ApproveReview(string) error
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

func (s *service) Get(id string) (User, error) {
	return s.store.Get(id)
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

func (s *service) GetReview(userId string) (UserReview, error) {
	return s.store.GetReview(userId)
}

type UpdateReviewRequest struct {
	UserId  string
	Comment string `schema:"comment"`
}

func (s *service) UpdateReview(req UpdateReviewRequest) error {
	userLog("UpdateReview req %+v", req)
	if len(req.Comment) > 100 {
		return errors.New("comment too long")
	}

	err := s.store.UpdateReview(UpdateReviewParams{
		UserId: req.UserId,
		Comment: sql.NullString{
			String: req.Comment,
			Valid:  req.Comment != "",
		},
	})
	return err
}

func (s *service) ListReviews() ([]UserReview, error) {
	return s.store.ListReviews()
}

func (s *service) ApproveReview(userId string) error {
	userLog("ApproveReview %s", userId)
    if userId == "" {
        return errors.New("invalid user")
    }
	return s.store.ApproveReview(userId)
}
