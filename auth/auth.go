package auth

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
)

type Service struct {
	oauthConf *oauth2.Config
}

func NewService(oauthConf *oauth2.Config) *Service {
	return &Service{
		oauthConf: oauthConf,
	}
}

func (s *Service) Routes(r *chi.Mux) {
	r.Route("/auth", func(r chi.Router) {
		r.Get("/login", s.getLoginPage)
		r.Get("/callback", s.getCallbackPage)
	})
}

func (s *Service) getLoginPage(w http.ResponseWriter, r *http.Request) {
    // TODO: state should be random
	fbLoginUrl := s.oauthConf.AuthCodeURL("state")
	http.Redirect(w, r, fbLoginUrl, http.StatusTemporaryRedirect)
}

func (s *Service) getCallbackPage(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Query().Get("state"))
	fmt.Println(r.URL.Query().Get("code"))

	w.Write([]byte("hi"))
}
