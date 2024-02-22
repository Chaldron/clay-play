package auth

import (
	"context"
	"errors"
	"fmt"
	"github/mattfan00/jvbe/config"
	userPkg "github/mattfan00/jvbe/user"
	"log"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

type Service interface {
    AuthCodeUrl(string) string
    InfoFromProvider(code string) (userPkg.ExternalUser, string, error)
    HandleLogin(code string) (userPkg.SessionUser, error)
}

type service struct {
	oidcProvider *oidc.Provider
	oauthConf    *oauth2.Config
	user         userPkg.Service
}

func NewService(conf *config.Config, user userPkg.Service) (*service, error) {
	provider, err := oidc.NewProvider(
		context.Background(),
		"https://"+conf.Oauth.Domain,
	)
	if err != nil {
		return &service{}, err
	}

	oauthConf := &oauth2.Config{
		ClientID:     conf.Oauth.ClientId,
		ClientSecret: conf.Oauth.ClientSecret,
		RedirectURL:  conf.OauthCallbackUrl(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		Endpoint:     provider.Endpoint(),
	}

	return &service{
		oidcProvider: provider,
		oauthConf:    oauthConf,
		user:         user,
	}, nil
}

func authLog(format string, s ...any) {
	log.Printf("auth/auth.go: %s", fmt.Sprintf(format, s...))
}

func (s *service) AuthCodeUrl(state string) string {
	return s.oauthConf.AuthCodeURL(state)
}

func (s *service) InfoFromProvider(code string) (userPkg.ExternalUser, string, error) {
	token, err := s.oauthConf.Exchange(context.Background(), code)
	if err != nil {
		return userPkg.ExternalUser{}, "", err
	}

	accessToken := token.AccessToken

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return userPkg.ExternalUser{}, "", errors.New("no id_token field in oauth2 token")
	}

	oidcConfig := &oidc.Config{
		ClientID: s.oauthConf.ClientID,
	}

	idToken, err := s.oidcProvider.Verifier(oidcConfig).Verify(context.Background(), rawIDToken)
	if err != nil {
		return userPkg.ExternalUser{}, "", err
	}

	var externalUser userPkg.ExternalUser
	err = idToken.Claims(&externalUser)
	if err != nil {
		return userPkg.ExternalUser{}, "", err
	}
	authLog("externalUser:%+v accessToken:%s", externalUser, accessToken)

	return externalUser, accessToken, nil
}

type customClaims struct {
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

func (s *service) HandleLogin(code string) (userPkg.SessionUser, error) {
	externalUser, accessToken, err := s.InfoFromProvider(code)
	if err != nil {
		return userPkg.SessionUser{}, err
	}

	u, err := s.user.HandleFromExternal(externalUser)
	if err != nil {
		return userPkg.SessionUser{}, err
	}

	sessionUser := u.ToSessionUser()

	// dont verify token since this should have come from oauth and I don't want to deal with verifying right now
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256"}))
	token, _, err := parser.ParseUnverified(accessToken, &customClaims{})
	if err != nil {
		return userPkg.SessionUser{}, err
	}

	claims, ok := token.Claims.(*customClaims)
	if !ok {
		return userPkg.SessionUser{}, errors.New("cannot get claims")
	}

	sessionUser.Permissions = claims.Permissions

	authLog("sessionUser:%+v", sessionUser)

	return sessionUser, nil
}
