package server

import (
	"net/http"
)

func (s *Server) spotifyGetPlaylists() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := s.Auth.GetClient(r)
		if err != nil {
			s.Auth.Redirect(w, r)
			return
		}

		usr, err := c.GetUser()
		if err != nil {
			s.log.Error("Failed to get user", err)
			s.internalServerError(w)
			return
		}
		pl, err := c.ListPlaylists(usr)
		if err != nil {
			s.log.Error("Failed to get playlists", err, "user_id", usr.ID)
			s.internalServerError(w)
			return
		}

		s.json(w, pl)
	}
}
