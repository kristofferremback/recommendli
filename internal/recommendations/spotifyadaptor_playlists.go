package recommendations

import (
	"context"
	"fmt"
	"sort"

	"github.com/kristofferostlund/recommendli/pkg/ctxhelper"
	"github.com/kristofferostlund/recommendli/pkg/spotifypaginator"
	"github.com/zmb3/spotify"
	"golang.org/x/sync/errgroup"
)

func (s *SpotifyAdaptor) ListPlaylists(ctx context.Context, userID string) ([]spotify.SimplePlaylist, error) {
	if err := ctxhelper.Closed(ctx); err != nil {
		return nil, fmt.Errorf("listing playlists for user %s: %w", userID, err)
	}
	playlists, err := s.listPlaylists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing playlists for user %s: %w", userID, err)
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
	storeKey := fmt.Sprintf("playlist_%s", playlistID)
	var stored spotify.FullPlaylist
	if exists, err := s.kv.Get(ctx, storeKey, &stored); err == nil && exists && stored.SnapshotID == snapshotID {
		return stored, nil
	} else if err != nil {
		return spotify.FullPlaylist{}, fmt.Errorf("getting playlist %s from store: %w", playlistID, err)
	}

	playlist, err := s.getPlaylist(ctx, playlistID)
	if err != nil {
		return spotify.FullPlaylist{}, err
	}

	if err := s.kv.Put(ctx, storeKey, playlist); err != nil {
		return spotify.FullPlaylist{}, fmt.Errorf("storing playlist %s: %w", playlistID, err)
	}

	return playlist, nil
}

func (s *SpotifyAdaptor) getPlaylist(ctx context.Context, playlistID string) (spotify.FullPlaylist, error) {
	p, err := s.spotify.GetPlaylist(spotify.ID(playlistID))
	if err != nil {
		return spotify.FullPlaylist{}, fmt.Errorf("getting playlist %s: %w", playlistID, err)
	}

	if len(p.Tracks.Tracks) < p.Tracks.Total {
		type indexAndTracks struct {
			index  int
			tracks []spotify.PlaylistTrack
		}

		paginator := spotifypaginator.New(spotifypaginator.InitialOffset(len(p.Tracks.Tracks)), spotifypaginator.Parallelism(10))
		itChan := make(chan indexAndTracks)
		g, ctx := errgroup.WithContext(ctx)
		g.Go(func() error {
			defer close(itChan)
			return paginator.Run(ctx, func(i int, opts *spotify.Options, next spotifypaginator.NextFunc) (result *spotifypaginator.NextResult, err error) {
				page, err := s.spotify.GetPlaylistTracksOpt(spotify.ID(p.ID), opts, "")
				if err != nil {
					return nil, err
				}
				itChan <- indexAndTracks{i, page.Tracks}
				s.log.Debug("listing playlist tracks", "playlist", p.Name, "counter", i, "offset", page.Offset, "total", page.Total)
				return next(page.Total), nil
			})
		})

		indexedTracks := make([]indexAndTracks, 0)
		for it := range itChan {
			indexedTracks = append(indexedTracks, it)
		}
		if err := g.Wait(); err != nil {
			return spotify.FullPlaylist{}, fmt.Errorf("listing tracks: %w", err)
		}
		sort.Slice(indexedTracks, func(i, j int) bool {
			return indexedTracks[i].index < indexedTracks[j].index
		})

		for _, it := range indexedTracks {
			p.Tracks.Tracks = append(p.Tracks.Tracks, it.tracks...)
		}
	}

	return *p, nil
}

func (s *SpotifyAdaptor) listPlaylists(ctx context.Context, userID string) ([]spotify.SimplePlaylist, error) {
	type indexAndPlaylists struct {
		index     int
		playlists []spotify.SimplePlaylist
	}

	paginator := spotifypaginator.New(spotifypaginator.Parallelism(10))
	ipChan := make(chan indexAndPlaylists)
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		defer close(ipChan)
		return paginator.Run(ctx, func(i int, opts *spotify.Options, next spotifypaginator.NextFunc) (*spotifypaginator.NextResult, error) {
			page, err := s.spotify.GetPlaylistsForUserOpt(userID, opts)
			if err != nil {
				return nil, err
			}
			ipChan <- indexAndPlaylists{i, page.Playlists}
			s.log.Debug("listing playlists for user", "user", userID, "counter", i, "offset", page.Offset, "total", page.Total)
			return next(page.Total), nil
		})
	})

	indexedPlaylists := make([]indexAndPlaylists, 0)
	for ip := range ipChan {
		indexedPlaylists = append(indexedPlaylists, ip)
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	sort.Slice(indexedPlaylists, func(i, j int) bool {
		return indexedPlaylists[i].index < indexedPlaylists[j].index
	})

	playlists := make([]spotify.SimplePlaylist, 0)
	for _, indexed := range indexedPlaylists {
		playlists = append(playlists, indexed.playlists...)
	}

	return playlists, nil
}
