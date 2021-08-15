package recommendations

import (
	"context"
	"fmt"

	"github.com/kristofferostlund/recommendli/pkg/ctxhelper"
	"github.com/kristofferostlund/recommendli/pkg/logging"
	"github.com/kristofferostlund/recommendli/pkg/spotifypaginator"
	"github.com/zmb3/spotify"
	"golang.org/x/sync/errgroup"
)

type KeyValueStore interface {
	Get(ctx context.Context, key string, out interface{}) (bool, error)
	Put(ctx context.Context, key string, data interface{}) error
}

const (
	kvPlaylistPrefix = "playlist"
)

type SpotifyAdaptor struct {
	spotify spotify.Client
	log     logging.Logger
	kv      KeyValueStore
}

type SpotifyAdaptorFactory struct {
	log   logging.Logger
	store KeyValueStore
}

func NewSpotifyProviderFactory(log logging.Logger, store KeyValueStore) *SpotifyAdaptorFactory {
	return &SpotifyAdaptorFactory{log: log, store: store}
}

func (f *SpotifyAdaptorFactory) New(spotifyClient spotify.Client) *SpotifyAdaptor {
	return &SpotifyAdaptor{spotify: spotifyClient, log: f.log, kv: f.store}
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

func (s *SpotifyAdaptor) GetPlaylist(ctx context.Context, playlistID string) (spotify.FullPlaylist, error) {
	if err := ctxhelper.Closed(ctx); err != nil {
		return spotify.FullPlaylist{}, fmt.Errorf("getting playlist %s: %w", playlistID, err)
	}
	return s.getStoredPlaylist(ctx, playlistID, "")
}

func (s *SpotifyAdaptor) PopulatePlaylists(ctx context.Context, simplePlaylists []spotify.SimplePlaylist) ([]spotify.FullPlaylist, error) {
	if err := ctxhelper.Closed(ctx); err != nil {
		return nil, fmt.Errorf("populating playlists: %w", err)
	}
	playlists := make([]spotify.FullPlaylist, 0, len(simplePlaylists))
	for _, p := range simplePlaylists {
		playlist, err := s.getStoredPlaylist(ctx, p.ID.String(), p.SnapshotID)
		if err != nil {
			return nil, err
		}
		playlists = append(playlists, playlist)
	}
	return playlists, nil
}

func (s *SpotifyAdaptor) getStoredPlaylist(ctx context.Context, playlistID, snapshotID string) (spotify.FullPlaylist, error) {
	cacheKey := fmt.Sprintf("playlist_%s", playlistID)
	var stored spotify.FullPlaylist
	if exists, err := s.kv.Get(ctx, cacheKey, &stored); err == nil && exists {
		s.log.Debug("returning stored playlist, snapshot ID matches stored value", "playlistID", playlistID, "snapshotID", snapshotID)
		return stored, nil
	} else if err != nil {
		return spotify.FullPlaylist{}, fmt.Errorf("getting playlist %s from store: %w", playlistID, err)
	}

	playlist, err := s.getPlaylist(ctx, playlistID)
	if err != nil {
		return spotify.FullPlaylist{}, err
	}

	if err := s.kv.Put(ctx, cacheKey, playlist); err != nil {
		return spotify.FullPlaylist{}, fmt.Errorf("storing playlist %s: %w", playlistID, err)
	}

	return playlist, nil
}

func (s *SpotifyAdaptor) getPlaylist(ctx context.Context, playlistID string) (spotify.FullPlaylist, error) {
	p, err := s.spotify.GetPlaylist(spotify.ID(playlistID))
	if err != nil {
		return spotify.FullPlaylist{}, fmt.Errorf("getting playlist %s: %w", playlistID, err)
	}
	playlist := *p
	if len(playlist.Tracks.Tracks) < playlist.Tracks.Total {
		paginator := spotifypaginator.New(
			spotifypaginator.PageSize(50),
			spotifypaginator.InitialOffset(len(playlist.Tracks.Tracks)),
			spotifypaginator.ProgressReporter(func(currentCount, totalCount, currentPage int) {
				if currentPage%2 == 0 || currentCount == totalCount {
					s.log.Debug("listing tracks for playlist", "playlist", playlist.Name, "count", currentCount, "total", totalCount, "page", currentPage)
				}
			}),
		)
		if err := paginator.Run(ctx, func(opts *spotify.Options, next spotifypaginator.NextFunc) (result *spotifypaginator.NextResult, err error) {
			page, err := s.spotify.GetPlaylistTracksOpt(spotify.ID(playlist.ID), opts, "")
			if err != nil {
				return nil, err
			}
			playlist.Tracks.Tracks = append(playlist.Tracks.Tracks, page.Tracks...)
			return next(page.Total), nil
		}); err != nil {
			return spotify.FullPlaylist{}, fmt.Errorf("listing tracks: %w", err)
		}
	}
	return playlist, nil
}

func (s *SpotifyAdaptor) listPlaylists(ctx context.Context, usr spotify.User) ([]spotify.SimplePlaylist, error) {
	playlists := make([]spotify.SimplePlaylist, 0)
	paginator := spotifypaginator.New(
		spotifypaginator.PageSize(50),
		spotifypaginator.ProgressReporter(func(currentCount, totalCount, currentPage int) {
			if currentPage%2 == 0 || currentCount == totalCount {
				s.log.Debug("listing playlists", "user", usr.DisplayName, "count", currentCount, "total", totalCount, "page", currentPage)
			}
		}),
	)
	playlistsChan := make(chan []spotify.SimplePlaylist)
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		defer close(playlistsChan)
		return paginator.RunAsync(ctx, func(opts *spotify.Options, next spotifypaginator.NextFunc) (*spotifypaginator.NextResult, error) {
			fmt.Println("offset", *opts.Offset)
			r, err := s.spotify.GetPlaylistsForUserOpt(usr.ID, opts)
			if err != nil {
				return nil, err
			}
			playlistsChan <- r.Playlists
			fmt.Println(r.Endpoint, "offset", r.Offset, "next", r.Next)
			return next(r.Total), nil
		})
	})

	for pls := range playlistsChan {
		playlists = append(playlists, pls...)
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return playlists, nil
}
