package auth

import (
	"github.com/mattfan00/jvbe/user"
)

type Service interface {
	GetExternalUser(code string) (user.ExternalUser, error)
}
