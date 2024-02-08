package auth

import (
	"context"
	"errors"
	"github/mattfan00/jvbe/config"
	"github/mattfan00/jvbe/user"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Service struct {
	oidcProvider *oidc.Provider
	oauthConf    *oauth2.Config
}

func NewService(conf *config.Config) (*Service, error) {
	provider, err := oidc.NewProvider(
		context.Background(),
		"https://"+conf.Oauth.Domain,
	)
	if err != nil {
		return &Service{}, err
	}

	oauthConf := &oauth2.Config{
		ClientID:     conf.Oauth.ClientId,
		ClientSecret: conf.Oauth.ClientSecret,
		RedirectURL:  conf.Oauth.CallbackUrl,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		Endpoint:     provider.Endpoint(),
	}

	return &Service{
		oidcProvider: provider,
		oauthConf:    oauthConf,
	}, nil
}

func (s *Service) AuthCodeUrl(state string) string {
	return s.oauthConf.AuthCodeURL(state)
}

func (s *Service) InfoFromProvider(code string) (user.ExternalUser, string, error) {
	token, err := s.oauthConf.Exchange(context.Background(), code)
	if err != nil {
		return user.ExternalUser{}, "", err
	}

	accessToken := token.AccessToken

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return user.ExternalUser{}, "", errors.New("no id_token field in oauth2 token")
	}

	oidcConfig := &oidc.Config{
		ClientID: s.oauthConf.ClientID,
	}

	idToken, err := s.oidcProvider.Verifier(oidcConfig).Verify(context.Background(), rawIDToken)
	if err != nil {
		return user.ExternalUser{}, "", err
	}

	var externalUser user.ExternalUser
	err = idToken.Claims(&externalUser)
	if err != nil {
		return user.ExternalUser{}, "", err
	}

	return externalUser, accessToken, nil
}
