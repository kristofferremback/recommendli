package server

import (
	"net/http"

	"github.com/kristofferostlund/recommendli/pkg/httphelpers"
)

func (s *Server) spotifyGetWhoAmI() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := s.Auth.GetClient(r)
		if err != nil {
			s.Auth.Redirect(w, r)
			return
		}

		usr, err := c.GetUser()
		if err != nil {
			s.log.Error("Failed to get user", err)
			httphelpers.InternalServerError(w)
			return
		}

		s.json(w, usr)
	}
}
