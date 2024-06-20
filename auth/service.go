package auth

import (
	"github.com/mattfan00/jvbe/logger"
	"github.com/mattfan00/jvbe/user"
)

type service struct {
	log logger.Logger
}

func NewService() (*service, error) {
	return &service{
		log: logger.NewNoopLogger(),
	}, nil
}

func (s *service) SetLogger(l logger.Logger) {
	s.log = l
}

func (s *service) GetExternalUser(code string) (user.ExternalUser, error) {
	//token, err := s.oauthConf.Exchange(context.Background(), code)
	//if err != nil {
	//	return user.ExternalUser{}, err
	//}
	//
	//accessToken := token.AccessToken
	//
	//rawIDToken, ok := token.Extra("id_token").(string)
	//if !ok {
	//	return user.ExternalUser{}, errors.New("no id_token field in oauth2 token")
	//}
	//
	//oidcConfig := &oidc.Config{
	//	ClientID: s.oauthConf.ClientID,
	//}
	//
	//idToken, err := s.oidcProvider.Verifier(oidcConfig).Verify(context.Background(), rawIDToken)
	//if err != nil {
	//	return user.ExternalUser{}, err
	//}
	//
	//var externalUser user.ExternalUser
	//err = idToken.Claims(&externalUser)
	//if err != nil {
	//	return user.ExternalUser{}, err
	//}
	//
	//fmt.Println("Token = ", accessToken)
	//
	//// dont verify token since this should have come from oauth and I don't want to deal with verifying right now
	//parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256"}))
	//accessTokenParsed, _, err := parser.ParseUnverified(accessToken, &customClaims{})
	//if err != nil {
	//	fmt.Println("err")
	//	return user.ExternalUser{}, err
	//}
	//
	//claims, ok := accessTokenParsed.Claims.(*customClaims)
	//fmt.Println("Claims = ", accessTokenParsed.Claims)
	//if !ok {
	//	return user.ExternalUser{}, errors.New("cannot get claims")
	//}
	//externalUser.Permissions = claims.Permissions
	//
	//s.log.Printf("externalUser:%+v accessToken:%s", externalUser, accessToken)

	return user.ExternalUser{}, nil
}
