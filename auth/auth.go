package auth

import (
	"fmt"
	"github/mattfan00/jvbe/facebook"
	userPkg "github/mattfan00/jvbe/user"
)

type Service struct {
	user     *userPkg.Service
	facebook *facebook.Service
}

func NewService(user *userPkg.Service, facebook *facebook.Service) *Service {
	return &Service{
		user:     user,
		facebook: facebook,
	}
}

// TODO: state should be random
var expectedStateVal = "state"

func (s *Service) ValidateState(state string) error {
	if state != expectedStateVal {
		err := fmt.Errorf("invalid oauth state, expected '%s', got '%s'", expectedStateVal, state)
		return err
	}

    return nil
}

func (s *Service) GetOauthLoginUrl() string {
	fbLoginUrl := s.facebook.GenerateAuthCodeUrl(expectedStateVal)
	return fbLoginUrl
}

func (s *Service) GetUserFromOauthCode(code string) (userPkg.User, error) {
	token, err := s.facebook.GetAccessToken(code)
	if err != nil {
		return userPkg.User{}, err
	}

	externalUser, err := s.facebook.GetUser(token)
	if err != nil {
		return userPkg.User{}, err
	}

	u, err := s.user.HandleFromExternal(externalUser)
	if err != nil {
		return userPkg.User{}, err
	}

	return u, nil
}
