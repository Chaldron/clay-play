package auth

import (
	"github.com/mattfan00/jvbe/user"
)

type Service interface {
	OAuthEnabled() bool
	AuthCodeUrl(string) string
	GetExternalUser(code string) (user.ExternalUser, error)
}
