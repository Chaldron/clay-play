package auth

import (
	"context"
	"errors"
	"log"

	"github.com/mattfan00/jvbe/config"
	"github.com/mattfan00/jvbe/user"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)
type service struct {
	oidcProvider *oidc.Provider
	oauthConf    *oauth2.Config
}

func NewService(conf *config.Config) (*service, error) {
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
	}, nil
}

func (s *service) AuthCodeUrl(state string) string {
	return s.oauthConf.AuthCodeURL(state)
}

type customClaims struct {
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

func (s *service) GetExternalUser(code string) (user.ExternalUser, error) {
	token, err := s.oauthConf.Exchange(context.Background(), code)
	if err != nil {
		return user.ExternalUser{}, err
	}

	accessToken := token.AccessToken

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return user.ExternalUser{}, errors.New("no id_token field in oauth2 token")
	}

	oidcConfig := &oidc.Config{
		ClientID: s.oauthConf.ClientID,
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

	// dont verify token since this should have come from oauth and I don't want to deal with verifying right now
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256"}))
	accessTokenParsed, _, err := parser.ParseUnverified(accessToken, &customClaims{})
	if err != nil {
		return user.ExternalUser{}, err
	}

	claims, ok := accessTokenParsed.Claims.(*customClaims)
	if !ok {
		return user.ExternalUser{}, errors.New("cannot get claims")
	}
	externalUser.Permissions = claims.Permissions

	log.Printf("externalUser:%+v accessToken:%s", externalUser, accessToken)

	return externalUser, nil
}
