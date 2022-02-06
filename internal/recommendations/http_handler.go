package recommendations

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/go-chi/chi"
	"github.com/kristofferostlund/recommendli/pkg/logging"
	"github.com/kristofferostlund/recommendli/pkg/sortby"
	"github.com/kristofferostlund/recommendli/pkg/srv"
	"github.com/zmb3/spotify"
)

const (
	playlistIDKey = "playlistID"
)

func NewRouter(svcFactory *ServiceFactory, spotifyProviderFactory *SpotifyAdaptorFactory, auth *AuthAdaptor, log logging.Logger) *chi.Mux {
	handler := &httpHandler{
		svcFactory:             svcFactory,
		spotifyProviderFactory: spotifyProviderFactory,
		auth:                   auth,
		log:                    log,
	}
	r := chi.NewRouter()

	ar := r.With(auth.Middleware())
	ar.Get("/v1/whoami", handler.withService(handler.whoami))
	ar.Get("/v1/check-current-track-in-library", handler.withService(handler.checkCurrentTrackInLibrary))
	ar.Get("/v1/generate-discovery-playlist", handler.withService(handler.generateDiscoveryPlaylist))
	ar.Get("/v1/album-for-current-track", handler.withService(handler.getAlbumForCurrentTrack))
	ar.Get("/v1/current-track", handler.withService(handler.getCurrentTrack))
	ar.Get("/v1/playlists", handler.withService(handler.listPlaylists))
	ar.Get("/v1/playlists/for", handler.withService(handler.getPlaylistMatchingPattern))
	ar.Get("/v1/playlists/{playlistID}", handler.withService(handler.getPlaylist))

	return r
}

type httpHandler struct {
	svcFactory             *ServiceFactory
	spotifyProviderFactory *SpotifyAdaptorFactory
	auth                   *AuthAdaptor
	log                    logging.Logger
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
		sHandler(h.svcFactory.New(h.spotifyProviderFactory.New(spotifyClient)))(w, r)
	}
}

func (h *httpHandler) whoami(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		usr, err := svc.GetCurrentUser(r.Context())
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
		playlists, err := svc.ListPlaylistsForCurrentUser(r.Context())
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

func (h *httpHandler) getPlaylistMatchingPattern(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pattern := r.URL.Query().Get("pattern")
		if pattern == "" {
			srv.JSONError(w, errors.New("pattern must be provided"), srv.Status(400))
			return
		}
		playlists, err := svc.GetCurrentUsersPlaylistMatchingPattern(r.Context(), pattern)
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

func (h *httpHandler) getPlaylist(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		playlistID := chi.URLParam(r, playlistIDKey)
		if playlistID == "" {
			srv.JSONError(w, errors.New("missing playlist ID in path"), srv.Status(400))
			return
		}
		playlist, err := svc.GetPlaylist(r.Context(), playlistID)
		if err != nil {
			h.log.Error("getting user's playlists", err)
			srv.InternalServerError(w)
			return
		}
		srv.JSON(w, playlist)
	}
}

func (h *httpHandler) checkCurrentTrackInLibrary(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentTrack, playlists, err := svc.CheckPlayingTrackInLibrary(r.Context())
		if err != nil && errors.As(err, &ErrNoCurrentTrack{}) {
			h.log.Error("user not listening to spotify", err)
			srv.JSONError(w, err, srv.Status(400))
			return
		} else if err != nil {
			h.log.Error("checking current track in library", err)
			srv.InternalServerError(w)
			return
		}
		srv.JSON(w, struct {
			InLibrary bool                     `json:"in_library"`
			Track     spotify.FullTrack        `json:"track"`
			Playlists []spotify.SimplePlaylist `json:"playlists"`
		}{Track: currentTrack, Playlists: playlists, InLibrary: len(playlists) > 0})
	}
}

func (h *httpHandler) generateDiscoveryPlaylist(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var playlist spotify.FullPlaylist
		var err error

		dryRunStr := strings.ToLower(r.URL.Query().Get("dryrun"))
		if dryRunStr == "true" {
			playlist, err = svc.DryRunDiscoveryPlaylist(r.Context())
		} else {
			playlist, err = svc.CreateDiscoveryPlaylist(r.Context())
		}

		if err != nil {
			h.log.Error("generating discovery playlist", err)
			srv.InternalServerError(w)
			return
		}
		srv.JSON(w, playlist)
	}
}

func (h *httpHandler) getAlbumForCurrentTrack(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		album, err := svc.GetCurrentlyPlayingTrackAlbum(r.Context())
		if err != nil && errors.As(err, &ErrNoCurrentTrack{}) {
			h.log.Error("user not listening to spotify", err)
			srv.JSONError(w, err, srv.Status(400))
			return
		} else if err != nil {
			h.log.Error("getting current track's album", err)
			srv.InternalServerError(w)
			return
		}
		srv.JSON(w, album)
	}
}

func (h *httpHandler) getCurrentTrack(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		track, isPlaying, err := svc.GetCurrentTrack(r.Context())
		if err != nil && errors.As(err, &ErrNoCurrentTrack{}) {
			h.log.Error("user not listening to spotify", err)
			srv.JSONError(w, err, srv.Status(400))
			return
		} else if err != nil {
			h.log.Error("getting current track's album", err)
			srv.InternalServerError(w)
			return
		}

		var ptrTrack *spotify.FullTrack
		if isPlaying {
			ptrTrack = &track
		}
		srv.JSON(w, struct {
			Track     *spotify.FullTrack `json:"track"`
			IsPlaying bool               `json:"is_playing"`
		}{ptrTrack, isPlaying})
	}
}
