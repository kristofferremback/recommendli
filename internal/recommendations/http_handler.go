package recommendations

import (
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/go-chi/chi"
	"github.com/kristofferostlund/recommendli/pkg/logging"
	"github.com/kristofferostlund/recommendli/pkg/sortby"
	"github.com/kristofferostlund/recommendli/pkg/srv"
)

func NewRouter(svcFactory *ServiceFactory, auth *AuthAdaptor, log logging.Logger) *chi.Mux {
	handler := &httpHandler{svcFactory: svcFactory, auth: auth, log: log}
	r := chi.NewRouter()

	ar := r.With(auth.Middleware())
	ar.Get("/v1/whoami", handler.withService(handler.whoami))
	ar.Get("/v1/playlists", handler.withService(handler.listPlaylists))

	return r
}

type httpHandler struct {
	svcFactory *ServiceFactory
	auth       *AuthAdaptor
	log        logging.Logger
}

type spotifyClientHandlerFunc func(svc *Service) http.HandlerFunc

func (h *httpHandler) withService(sHandler spotifyClientHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		spotifyClient, err := h.auth.GetClient(r)
		if err != nil && errors.Is(err, NoAuthenticationError) {
			srv.JSONError(w, fmt.Errorf("user not signed in: %w", err), srv.Status(http.StatusUnauthorized))
		} else if err != nil {
			h.log.Error("getting spotify client", err)
			srv.InternalServerError(w)
			return
		}
		sHandler(h.svcFactory.New(spotifyClient))(w, r)
	}
}

func (h *httpHandler) whoami(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		usr, err := svc.currentUser(r.Context())
		if err != nil {
			h.log.Error("getting current user", err)
			srv.InternalServerError(w)
			return
		}
		srv.JSON(w, usr)
	}
}

func (h *httpHandler) listPlaylists(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		usr, err := svc.currentUser(r.Context())
		if err != nil {
			h.log.Error("getting current user", err)
			srv.InternalServerError(w)
			return
		}
		playlists, err := svc.listPlaylists(r.Context(), usr)
		if err != nil {
			h.log.Error("getting user's playlists", err)
			srv.InternalServerError(w)
			return
		}
		sort.Slice(playlists, func(i, j int) bool {
			return sortby.PaddedNumbers(playlists[i].Name, playlists[j].Name, 10, true)
		})
		srv.JSON(w, playlists)
	}
}