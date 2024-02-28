package user

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type Service interface {
	HandleFromExternal(ExternalUser) (User, error)
    GetReview(string) (UserReview, error)
	UpdateReview(UpdateReviewRequest) error
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

func (s *service) GetReview(userId string) (UserReview, error) {
    return s.store.GetReview(userId)
}

type UpdateReviewRequest struct {
	UserId  string
	Comment string `schema:"comment"`
}

func (s *service) UpdateReview(req UpdateReviewRequest) error {
	userLog("UpdateReview req %+v", req)
	if len(req.Comment) > 500 {
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

