package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (s *Server) setupRoutes(r *chi.Mux) {
	r.Get("/status", getStatus())
	r.Method(http.MethodGet, "/metrics", promhttp.Handler())

	r.Get("/v1/spotify/auth/callback", s.Auth.TokenCallbackHandler())

	ar := r.With(s.Auth.Middleware())
	ar.Get("/v1/spotify/whoami", s.spotifyGetWhoAmI())
	ar.Get("/v1/spotify/playlists", s.spotifyGetPlaylists())
}
