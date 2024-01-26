package auth

import (
	"fmt"
	"github/mattfan00/jvbe/facebook"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Service struct {
	facebook *facebook.Service
}

func New(facebook *facebook.Service) *Service {
	return &Service{
		facebook: facebook,
	}
}

func (s *Service) Routes(r *chi.Mux) {
	r.Route("/auth", func(r chi.Router) {
		r.Get("/login", s.getLoginPage)
		r.Get("/callback", s.getCallbackPage)
	})
}

var expectedStateVal = "state"

func (s *Service) getLoginPage(w http.ResponseWriter, r *http.Request) {
	// TODO: state should be random
	fbLoginUrl := s.facebook.GenerateAuthCodeUrl(expectedStateVal)
	http.Redirect(w, r, fbLoginUrl, http.StatusTemporaryRedirect)
}

func (s *Service) getCallbackPage(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if state != expectedStateVal {
		err := fmt.Errorf("invalid oauth state, expected '%s', got '%s'", expectedStateVal, state)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// now that we are succesfully authenticated, use the code we got back to get the access token
	code := r.URL.Query().Get("code")
	token, err := s.facebook.GetAccessToken(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := s.facebook.GetUser(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(fmt.Sprintf("%+v", user)))
}
