package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/mattfan00/jvbe/config"
	"github.com/mattfan00/jvbe/logger"
	"github.com/mattfan00/jvbe/user"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

type service struct {
	oidcProvider *oidc.Provider
	oauthConf    *oauth2.Config
	log          logger.Logger
}

func NewService(conf *config.Config) (*service, error) {
	provider, err := oidc.NewProvider(
		context.Background(),
		"https://"+conf.Oauth.Domain,
	)
	if err != nil {
		return &service{}, nil
	}

	oauthConf := &oauth2.Config{
		ClientID:     conf.Oauth.ClientId,
		ClientSecret: conf.Oauth.ClientSecret,
		RedirectURL:  conf.OauthCallbackUrl(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		Endpoint:     provider.Endpoint(),
	}

	fmt.Print(oauthConf)

	return &service{
		oidcProvider: provider,
		oauthConf:    oauthConf,
		log:          logger.NewNoopLogger(),
	}, nil
}

func (s *service) SetLogger(l logger.Logger) {
	s.log = l
}

func (s *service) OAuthEnabled() bool {
	return s.oauthConf != nil
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

	fmt.Println("Token = ", accessToken)

	// dont verify token since this should have come from oauth and I don't want to deal with verifying right now
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256"}))
	accessTokenParsed, _, err := parser.ParseUnverified(accessToken, &customClaims{})
	if err != nil {
		fmt.Println("err")
		return user.ExternalUser{}, err
	}

	claims, ok := accessTokenParsed.Claims.(*customClaims)
	fmt.Println("Claims = ", accessTokenParsed.Claims)
	if !ok {
		return user.ExternalUser{}, errors.New("cannot get claims")
	}
	externalUser.Permissions = claims.Permissions

	s.log.Printf("externalUser:%+v accessToken:%s", externalUser, accessToken)

	return externalUser, nil
}
