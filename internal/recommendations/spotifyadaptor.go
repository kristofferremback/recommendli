package recommendations

import (
	"context"
	"fmt"

	"github.com/kristofferostlund/recommendli/pkg/ctxhelper"
	"github.com/kristofferostlund/recommendli/pkg/logging"
	"github.com/kristofferostlund/recommendli/pkg/spotifypaginator"
	"github.com/zmb3/spotify"
)

type KeyValueStore interface {
	Get(ctx context.Context, key string, out interface{}) error
	Put(ctx context.Context, key string, data interface{}) error
}

type SpotifyAdaptor struct {
	spotify spotify.Client
	log     logging.Logger
	store   KeyValueStore
}

type SpotifyAdaptorFactory struct {
	log   logging.Logger
	store KeyValueStore
}

func NewSpotifyProviderFactory(log logging.Logger, store KeyValueStore) *SpotifyAdaptorFactory {
	return &SpotifyAdaptorFactory{log: log, store: store}
}

func (f *SpotifyAdaptorFactory) New(spotifyClient spotify.Client) *SpotifyAdaptor {
	return &SpotifyAdaptor{spotify: spotifyClient, log: f.log, store: f.store}
}

func (s *SpotifyAdaptor) ListPlaylists(ctx context.Context, usr spotify.User) ([]spotify.SimplePlaylist, error) {
	if err := ctxhelper.Closed(ctx); err != nil {
		return nil, fmt.Errorf("listing playlists for user %s: %w", usr.ID, err)
	}
	playlists, err := s.listPlaylists(ctx, usr)
	if err != nil {
		return nil, fmt.Errorf("listing playlists for user %s: %w", usr.ID, err)
	}
	return playlists, nil
}

func (s *SpotifyAdaptor) listPlaylists(ctx context.Context, usr spotify.User) ([]spotify.SimplePlaylist, error) {
	playlists := make([]spotify.SimplePlaylist, 0)
	paginator := spotifypaginator.New(
		spotifypaginator.PageSize(50),
		spotifypaginator.ProgressReporter(func(currentCount, totalCount, currentPage int) {
			if currentPage%2 == 0 || currentCount == totalCount {
				s.log.Info("listing playlists", "user", usr.DisplayName, "count", currentCount, "total", totalCount, "page", currentPage)
			}
		}),
	)
	if err := paginator.Run(ctx, func(opts *spotify.Options, next spotifypaginator.NextFunc) (*spotifypaginator.NextResult, error) {
		r, err := s.spotify.GetPlaylistsForUserOpt(usr.ID, opts)
		if err != nil {
			return nil, err
		}
		playlists = append(playlists, r.Playlists...)
		return next(len(playlists), r.Total), nil
	}); err != nil {
		return nil, err
	}
	return playlists, nil
}

func (s *SpotifyAdaptor) CurrentUser(ctx context.Context) (spotify.User, error) {
	if err := ctxhelper.Closed(ctx); err != nil {
		return spotify.User{}, fmt.Errorf("getting current user: %w", err)
	}
	usr, err := s.spotify.CurrentUser()
	if err != nil {
		return spotify.User{}, fmt.Errorf("getting current user: %w", err)
	}
	return usr.User, nil
}
