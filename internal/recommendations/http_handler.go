package recommendations

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/kristofferostlund/recommendli/pkg/slogutil"
	"github.com/kristofferostlund/recommendli/pkg/sortby"
	"github.com/kristofferostlund/recommendli/pkg/srv"
	"github.com/zmb3/spotify"
)

const (
	playlistIDKey = "playlistID"
)

func NewRouter(svcFactory *ServiceFactory, spotifyProviderFactory *SpotifyAdaptorFactory, auth *AuthAdaptor) *chi.Mux {
	handler := &httpHandler{
		svcFactory:             svcFactory,
		spotifyProviderFactory: spotifyProviderFactory,
		auth:                   auth,
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
	ar.Get("/v1/index/summary", handler.withService(handler.getIndexSummary))

	return r
}

type httpHandler struct {
	svcFactory             *ServiceFactory
	spotifyProviderFactory *SpotifyAdaptorFactory
	auth                   *AuthAdaptor
}

type spotifyClientHandlerFunc func(svc *Service) http.HandlerFunc

func (h *httpHandler) withService(sHandler spotifyClientHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		spotifyClient, err := h.auth.GetClient(r)
		if err != nil && errors.Is(err, ErrNoAuthentication) {
			srv.JSONError(w, fmt.Errorf("user not signed in: %w", err), srv.Status(http.StatusUnauthorized))
		} else if err != nil {
			slog.ErrorContext(ctx, "getting spotify client", slogutil.Error(err))
			srv.InternalServerError(w)
			return
		}
		sHandler(h.svcFactory.New(h.spotifyProviderFactory.New(spotifyClient)))(w, r)
	}
}

func (h *httpHandler) whoami(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		usr, err := svc.GetCurrentUser(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "getting current user", slogutil.Error(err))
			srv.InternalServerError(w)
			return
		}
		srv.JSON(w, usr)
	}
}

func (h *httpHandler) listPlaylists(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		playlists, err := svc.ListPlaylistsForCurrentUser(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "getting user's playlists", slogutil.Error(err))
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
		ctx := r.Context()
		pattern := r.URL.Query().Get("pattern")
		if pattern == "" {
			srv.JSONError(w, errors.New("pattern must be provided"), srv.Status(400))
			return
		}
		playlists, err := svc.GetCurrentUsersPlaylistMatchingPattern(ctx, pattern)
		if err != nil {
			slog.ErrorContext(ctx, "getting user's playlists", slogutil.Error(err))
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
		ctx := r.Context()
		playlistID := chi.URLParam(r, playlistIDKey)
		if playlistID == "" {
			srv.JSONError(w, errors.New("missing playlist ID in path"), srv.Status(400))
			return
		}
		playlist, err := svc.GetPlaylist(ctx, playlistID)
		if err != nil {
			slog.ErrorContext(ctx, "getting user's playlists", slogutil.Error(err))
			srv.InternalServerError(w)
			return
		}
		srv.JSON(w, playlist)
	}
}

func (h *httpHandler) checkCurrentTrackInLibrary(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		currentTrack, playlists, err := svc.CheckPlayingTrackInLibrary(ctx)
		if err != nil && errors.As(err, &ErrNoCurrentTrack{}) {
			slog.ErrorContext(ctx, "user not listening to spotify", slogutil.Error(err))
			srv.JSONError(w, err, srv.Status(400))
			return
		} else if err != nil {
			slog.ErrorContext(ctx, "checking current track in library", slogutil.Error(err))
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
		ctx := r.Context()
		var playlist spotify.FullPlaylist
		var err error

		dryRunStr := strings.ToLower(r.URL.Query().Get("dryrun"))
		if dryRunStr == "true" {
			playlist, err = svc.DryRunDiscoveryPlaylist(ctx)
		} else {
			playlist, err = svc.CreateDiscoveryPlaylist(ctx)
		}

		if err != nil {
			slog.ErrorContext(ctx, "generating discovery playlist", slogutil.Error(err))
			srv.InternalServerError(w)
			return
		}
		srv.JSON(w, playlist)
	}
}

func (h *httpHandler) getAlbumForCurrentTrack(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		album, err := svc.GetCurrentlyPlayingTrackAlbum(ctx)
		if err != nil && errors.As(err, &ErrNoCurrentTrack{}) {
			slog.ErrorContext(ctx, "user not listening to spotify", slogutil.Error(err))
			srv.JSONError(w, err, srv.Status(400))
			return
		} else if err != nil {
			slog.ErrorContext(ctx, "getting current track's album", slogutil.Error(err))
			srv.InternalServerError(w)
			return
		}
		srv.JSON(w, album)
	}
}

func (h *httpHandler) getCurrentTrack(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		track, isPlaying, err := svc.GetCurrentTrack(ctx)
		if err != nil && errors.As(err, &ErrNoCurrentTrack{}) {
			slog.ErrorContext(ctx, "user not listening to spotify", slogutil.Error(err))
			srv.JSONError(w, err, srv.Status(400))
			return
		} else if err != nil {
			slog.ErrorContext(ctx, "getting current track's album", slogutil.Error(err))
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

func (h *httpHandler) getIndexSummary(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		summary, err := svc.GetIndexSummary(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "starting indexing", slogutil.Error(err))
			srv.InternalServerError(w)
			return
		}
		srv.JSON(w, struct {
			Playlists    int `json:"playlists"`
			UniqueTracks int `json:"unique_tracks"`
		}{Playlists: summary.Playlists, UniqueTracks: summary.UniqueTracks})
	}
}
