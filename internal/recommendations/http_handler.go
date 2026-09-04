package recommendations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/kristofferostlund/recommendli/pkg/slogutil"
	"github.com/kristofferostlund/recommendli/pkg/srv"
	"github.com/zmb3/spotify"
)

const (
	playlistIDKey = "playlistID"
	trackIDKey    = "trackID"
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
	ar.Post("/v1/index/sync", handler.withService(handler.syncIndex))
	ar.Get("/v1/tracks/{trackID}/library-status", handler.withService(handler.getTrackLibraryStatus))
	ar.Get("/v1/playback", handler.withService(handler.getPlayback))
	ar.Get("/v1/playback/queue", handler.withService(handler.getPlaybackQueue))
	ar.Get("/v1/playback/history", handler.withService(handler.getPlaybackHistory))
	ar.Post("/v1/playback/play", handler.withService(handler.play))
	ar.Post("/v1/playback/pause", handler.withService(handler.pause))
	ar.Post("/v1/playback/next", handler.withService(handler.next))
	ar.Post("/v1/playback/previous", handler.withService(handler.previous))
	ar.Post("/v1/playback/seek", handler.withService(handler.seek))
	ar.Post("/v1/playback/queue/skip", handler.withService(handler.skipPlaybackQueue))

	return r
}

type httpHandler struct {
	svcFactory             *ServiceFactory
	spotifyProviderFactory *SpotifyAdaptorFactory
	auth                   *AuthAdaptor
}

type Service interface {
	CheckPlayingTrackInLibrary(ctx context.Context) (spotify.FullTrack, []spotify.SimplePlaylist, error)
	CreateDiscoveryPlaylist(ctx context.Context) (spotify.FullPlaylist, error)
	DryRunDiscoveryPlaylist(ctx context.Context) (spotify.FullPlaylist, error)
	GetCurrentlyPlayingTrackAlbum(ctx context.Context) (spotify.FullAlbum, error)
	GetCurrentTrack(ctx context.Context) (spotify.FullTrack, bool, error)
	GetCurrentUser(ctx context.Context) (spotify.User, error)
	GetCurrentUsersPlaylistMatchingPattern(ctx context.Context, pattern string) ([]spotify.FullPlaylist, error)
	GetIndexSummary(ctx context.Context) (IndexSummary, error)
	SyncIndex(ctx context.Context) (IndexSummary, error)
	LookupTrackInLibrary(ctx context.Context, trackID string) ([]spotify.SimplePlaylist, error)
	GetPlayback(ctx context.Context) (Playback, error)
	GetPlaybackQueue(ctx context.Context) (PlaybackQueue, error)
	GetPlaybackHistory(ctx context.Context, limit int) ([]spotify.RecentlyPlayedItem, error)
	Play(ctx context.Context, trackID string) error
	Pause(ctx context.Context) error
	Next(ctx context.Context) error
	Previous(ctx context.Context) error
	Seek(ctx context.Context, positionMs int) error
	SkipPlaybackQueue(ctx context.Context, req QueueSkipRequest) error
	GetPlaylist(ctx context.Context, playlistID string) (spotify.FullPlaylist, error)
	ListPlaylistsForCurrentUser(ctx context.Context) ([]spotify.SimplePlaylist, error)
}

type spotifyClientHandlerFunc func(svc Service) http.HandlerFunc

func (h *httpHandler) withService(sHandler spotifyClientHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		spotifyClient, err := h.auth.GetClient(r)
		if err != nil && errors.Is(err, ErrNoAuthentication) {
			srv.JSONError(w, fmt.Errorf("user not signed in: %w", err), srv.Status(http.StatusUnauthorized))
		} else if err != nil {
			slog.ErrorContext(ctx, "getting spotify client", slogutil.Error(err))
			srv.InternalServerError(w, err)
			return
		}
		httpClient, err := h.auth.GetHTTPClient(r)
		if err != nil {
			slog.ErrorContext(ctx, "getting Spotify HTTP client", slogutil.Error(err))
			srv.InternalServerError(w, err)
			return
		}
		sHandler(h.svcFactory.New(h.spotifyProviderFactory.New(spotifyClient, httpClient)))(w, r)
	}
}

func (h *httpHandler) whoami(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		usr, err := svc.GetCurrentUser(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "getting current user", slogutil.Error(err))
			srv.InternalServerError(w, err)
			return
		}
		srv.JSON(w, usr)
	}
}

func (h *httpHandler) listPlaylists(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		playlists, err := svc.ListPlaylistsForCurrentUser(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "getting user's playlists", slogutil.Error(err))
			srv.InternalServerError(w, err)
			return
		}
		srv.JSON(w, playlists)
	}
}

func (h *httpHandler) getPlaylistMatchingPattern(svc Service) http.HandlerFunc {
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
			srv.InternalServerError(w, err)
			return
		}
		srv.JSON(w, playlists)
	}
}

func (h *httpHandler) getPlaylist(svc Service) http.HandlerFunc {
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
			srv.InternalServerError(w, err)
			return
		}
		srv.JSON(w, playlist)
	}
}

func (h *httpHandler) checkCurrentTrackInLibrary(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		currentTrack, playlists, err := svc.CheckPlayingTrackInLibrary(ctx)
		if err != nil && errors.As(err, &ErrNoCurrentTrack{}) {
			slog.ErrorContext(ctx, "user not listening to spotify", slogutil.Error(err))
			srv.JSONError(w, err, srv.Status(400))
			return
		} else if err != nil {
			slog.ErrorContext(ctx, "checking current track in library", slogutil.Error(err))
			srv.InternalServerError(w, err)
			return
		}
		srv.JSON(w, struct {
			InLibrary bool                     `json:"in_library"`
			Track     spotify.FullTrack        `json:"track"`
			Playlists []spotify.SimplePlaylist `json:"playlists"`
		}{Track: currentTrack, Playlists: playlists, InLibrary: len(playlists) > 0})
	}
}

func (h *httpHandler) generateDiscoveryPlaylist(svc Service) http.HandlerFunc {
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
			srv.InternalServerError(w, err)
			return
		}
		srv.JSON(w, playlist)
	}
}

func (h *httpHandler) getAlbumForCurrentTrack(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		album, err := svc.GetCurrentlyPlayingTrackAlbum(ctx)
		if err != nil && errors.As(err, &ErrNoCurrentTrack{}) {
			slog.ErrorContext(ctx, "user not listening to spotify", slogutil.Error(err))
			srv.JSONError(w, err, srv.Status(400))
			return
		} else if err != nil {
			slog.ErrorContext(ctx, "getting current track's album", slogutil.Error(err))
			srv.InternalServerError(w, err)
			return
		}
		srv.JSON(w, album)
	}
}

func (h *httpHandler) getCurrentTrack(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		track, isPlaying, err := svc.GetCurrentTrack(ctx)
		if err != nil && errors.As(err, &ErrNoCurrentTrack{}) {
			slog.ErrorContext(ctx, "user not listening to spotify", slogutil.Error(err))
			srv.JSONError(w, err, srv.Status(400))
			return
		} else if err != nil {
			slog.ErrorContext(ctx, "getting current track's album", slogutil.Error(err))
			srv.InternalServerError(w, err)
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

func (h *httpHandler) getIndexSummary(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		summary, err := svc.GetIndexSummary(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "getting index summary", slogutil.Error(err))
			srv.InternalServerError(w, err)
			return
		}
		srv.JSON(w, summary)
	}
}

func (h *httpHandler) syncIndex(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		summary, err := svc.SyncIndex(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "syncing index", slogutil.Error(err))
			srv.InternalServerError(w, err)
			return
		}
		srv.JSON(w, summary)
	}
}

func (h *httpHandler) getTrackLibraryStatus(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		trackID := chi.URLParam(r, trackIDKey)
		if trackID == "" {
			srv.JSONError(w, errors.New("missing track ID in path"), srv.Status(http.StatusBadRequest))
			return
		}
		playlists, err := svc.LookupTrackInLibrary(r.Context(), trackID)
		if err != nil {
			slog.ErrorContext(r.Context(), "looking up track in library", slogutil.Error(err))
			srv.InternalServerError(w, err)
			return
		}
		srv.JSON(w, struct {
			InLibrary bool                     `json:"in_library"`
			Playlists []spotify.SimplePlaylist `json:"playlists"`
		}{InLibrary: len(playlists) > 0, Playlists: playlists})
	}
}

func (h *httpHandler) getPlayback(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		playback, err := svc.GetPlayback(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "getting playback", slogutil.Error(err))
			srv.InternalServerError(w, err)
			return
		}
		srv.JSON(w, playback)
	}
}

func (h *httpHandler) getPlaybackQueue(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queue, err := svc.GetPlaybackQueue(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "getting playback queue", slogutil.Error(err))
			srv.InternalServerError(w, err)
			return
		}
		srv.JSON(w, queue)
	}
}

func (h *httpHandler) getPlaybackHistory(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 20
		if value := r.URL.Query().Get("limit"); value != "" {
			parsed, err := strconv.Atoi(value)
			if err != nil || parsed < 1 || parsed > 50 {
				srv.JSONError(w, errors.New("limit must be between 1 and 50"), srv.Status(http.StatusBadRequest))
				return
			}
			limit = parsed
		}
		items, err := svc.GetPlaybackHistory(r.Context(), limit)
		if err != nil {
			slog.ErrorContext(r.Context(), "getting playback history", slogutil.Error(err))
			srv.InternalServerError(w, err)
			return
		}
		srv.JSON(w, items)
	}
}

func (h *httpHandler) play(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			TrackID string `json:"track_id"`
		}
		if err := decodeOptionalJSON(r, &body); err != nil {
			srv.JSONError(w, err, srv.Status(http.StatusBadRequest))
			return
		}
		h.runPlaybackCommand(w, r, "starting playback", func(ctx context.Context) error {
			return svc.Play(ctx, body.TrackID)
		})
	}
}

func (h *httpHandler) pause(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.runPlaybackCommand(w, r, "pausing playback", svc.Pause)
	}
}

func (h *httpHandler) next(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.runPlaybackCommand(w, r, "skipping to next track", svc.Next)
	}
}

func (h *httpHandler) previous(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.runPlaybackCommand(w, r, "skipping to previous track", svc.Previous)
	}
}

func (h *httpHandler) seek(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			PositionMs int `json:"position_ms"`
		}
		if err := decodeJSON(r, &body); err != nil {
			srv.JSONError(w, err, srv.Status(http.StatusBadRequest))
			return
		}
		if body.PositionMs < 0 {
			srv.JSONError(w, errors.New("position_ms cannot be negative"), srv.Status(http.StatusBadRequest))
			return
		}
		h.runPlaybackCommand(w, r, "seeking playback", func(ctx context.Context) error {
			return svc.Seek(ctx, body.PositionMs)
		})
	}
}

func (h *httpHandler) skipPlaybackQueue(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body QueueSkipRequest
		if err := decodeJSON(r, &body); err != nil {
			srv.JSONError(w, err, srv.Status(http.StatusBadRequest))
			return
		}
		if body.Position < 0 || body.ExpectedTrackID == "" {
			srv.JSONError(w, errors.New("position and expected_track_id are required"), srv.Status(http.StatusBadRequest))
			return
		}
		if err := svc.SkipPlaybackQueue(r.Context(), body); err != nil {
			if errors.Is(err, ErrQueueChanged) {
				srv.JSONError(w, err, srv.Status(http.StatusConflict))
				return
			}
			slog.ErrorContext(r.Context(), "skipping playback queue", slogutil.Error(err))
			srv.InternalServerError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

type playbackCommand func(context.Context) error

func (h *httpHandler) runPlaybackCommand(w http.ResponseWriter, r *http.Request, action string, command playbackCommand) {
	if err := command(r.Context()); err != nil {
		slog.ErrorContext(r.Context(), action, slogutil.Error(err))
		srv.InternalServerError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeJSON(r *http.Request, out any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return fmt.Errorf("invalid JSON body: %w", err)
	}
	return nil
}

func decodeOptionalJSON(r *http.Request, out any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("invalid JSON body: %w", err)
	}
	return nil
}
