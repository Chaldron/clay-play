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
	oauth2Conf   *oauth2.Config
}

func NewService(conf *config.Config) (*Service, error) {
	provider, err := oidc.NewProvider(
		context.Background(),
		"https://"+conf.Auth0.Domain+"/",
	)
	if err != nil {
		return &Service{}, err
	}

	oauth2Conf := &oauth2.Config{
		ClientID:     conf.Auth0.ClientId,
		ClientSecret: conf.Auth0.ClientSecret,
		RedirectURL:  conf.Auth0.CallbackUrl,
		Scopes:       []string{oidc.ScopeOpenID, "profile"},
		Endpoint:     provider.Endpoint(),
	}

	return &Service{
		oidcProvider: provider,
		oauth2Conf:   oauth2Conf,
	}, nil
}

func (s *Service) AuthCodeUrl(state string) string {
	return s.oauth2Conf.AuthCodeURL(state)
}

func (s *Service) ExternalUserFromProvider(code string) (user.ExternalUser, error) {
	token, err := s.oauth2Conf.Exchange(context.Background(), code)
	if err != nil {
		return user.ExternalUser{}, err
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return user.ExternalUser{}, errors.New("no id_token field in oauth2 token")
	}

	oidcConfig := &oidc.Config{
		ClientID: s.oauth2Conf.ClientID,
	}

	idToken, err := s.oidcProvider.Verifier(oidcConfig).Verify(context.Background(), rawIDToken)
	if err != nil {
		return user.ExternalUser{}, err
	}

	var externalUser user.ExternalUser
	err = idToken.Claims(&externalUser)
	if err != nil {
		return user.ExternalUser{}, err
	}

	return externalUser, nil
}
