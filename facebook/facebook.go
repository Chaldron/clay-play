package facebook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
)

type Service struct {
	oauthConf *oauth2.Config
}

func New(oauthConf *oauth2.Config) *Service {
	return &Service{
		oauthConf: oauthConf,
	}
}

func (s *Service) GenerateAuthCodeUrl(state string) string {
	return s.oauthConf.AuthCodeURL(state)
}

func (s *Service) GetAccessToken(authCode string) (string, error) {
	token, err := s.oauthConf.Exchange(context.Background(), authCode)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

type User struct {
	Id       string `json:"id"`
	FullName string `json:"name"`
}

func (s *Service) GetUser(accessToken string) (User, error) {
	res, err := http.Get("https://graph.facebook.com/v19.0/me?fields=id,name&access_token=" + url.QueryEscape(accessToken))
	if err != nil {
		return User{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return User{}, fmt.Errorf("cannot get user info")
	}

	var user User
	err = json.NewDecoder(res.Body).Decode(&user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}
