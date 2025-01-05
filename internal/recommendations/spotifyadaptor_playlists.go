package recommendations

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/kristofferostlund/recommendli/pkg/ctxhelper"
	"github.com/kristofferostlund/recommendli/pkg/paginator"
	"github.com/kristofferostlund/recommendli/pkg/spotifyutil"
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

func (s *SpotifyAdaptor) CreatePlaylist(ctx context.Context, userID, name string, trackIDs []string) (spotify.FullPlaylist, error) {
	if err := ctxhelper.Closed(ctx); err != nil {
		return spotify.FullPlaylist{}, fmt.Errorf("creating playlist %s for user %s: %w", name, userID, err)
	}
	created, err := s.spotify.CreatePlaylistForUser(userID, name, "Auto generated playlist by recommendli", true)
	if err != nil {
		return spotify.FullPlaylist{}, fmt.Errorf("creating playlist %s for user %s: %w", name, userID, err)
	}

	if err := s.addTracksToPlaylist(ctx, created.ID.String(), trackIDs); err != nil {
		return spotify.FullPlaylist{}, err
	}
	return s.GetPlaylist(ctx, created.ID.String())
}

func (s *SpotifyAdaptor) TruncatePlaylist(ctx context.Context, playlistID string, snapshotID string) error {
	if err := ctxhelper.Closed(ctx); err != nil {
		return fmt.Errorf("truncating playlist %s: %w", playlistID, err)
	}

	playlist, err := s.getStoredPlaylist(ctx, playlistID, snapshotID)
	if err != nil {
		return fmt.Errorf("getting existing playlist when truncating playlist %s: %w", playlistID, err)
	}

	if playlist.Tracks.Total > 0 {
		prevTrackIDs := make([]spotify.ID, 0)
		for _, t := range playlist.Tracks.Tracks {
			prevTrackIDs = append(prevTrackIDs, t.Track.ID)
		}
		if _, err := s.spotify.RemoveTracksFromPlaylist(spotify.ID(playlistID), prevTrackIDs...); err != nil {
			return fmt.Errorf("removing tracks from playlist %s: %w", playlistID, err)
		}
	}

	return nil
}

func (s *SpotifyAdaptor) SetPlaylistTracks(ctx context.Context, playlistID string, trackIDs []string) (spotify.FullPlaylist, error) {
	if err := ctxhelper.Closed(ctx); err != nil {
		return spotify.FullPlaylist{}, fmt.Errorf("seting tracks for playlist %s: %w", playlistID, err)
	}
	if err := s.addTracksToPlaylist(ctx, playlistID, trackIDs); err != nil {
		return spotify.FullPlaylist{}, err
	}
	return s.getPlaylist(ctx, playlistID)
}

func (s *SpotifyAdaptor) addTracksToPlaylist(ctx context.Context, id string, trackIDs []string) error {
	pgtr := paginator.New(paginator.PageSize(100), paginator.InitialTotalCount(len(trackIDs)))
	if err := pgtr.RunSync(ctx, func(index int, opts paginator.PageOpts, next paginator.NextFunc) (result *paginator.NextResult, err error) {
		from, to := opts.Offset, opts.Offset+opts.Limit
		spotifyIDs := make([]spotify.ID, 0)
		for _, id := range trackIDs[from:to] {
			spotifyIDs = append(spotifyIDs, spotify.ID(id))
		}
		if _, err := s.spotify.AddTracksToPlaylist(spotify.ID(id), spotifyIDs...); err != nil {
			return nil, err
		}
		return next(len(trackIDs)), nil
	}); err != nil {
		return fmt.Errorf("adding tracks to playlist %s: %w", id, err)
	}
	return nil
}

func (s *SpotifyAdaptor) PopulatePlaylists(ctx context.Context, simplePlaylists []spotify.SimplePlaylist) ([]spotify.FullPlaylist, error) {
	if err := ctxhelper.Closed(ctx); err != nil {
		return nil, fmt.Errorf("populating playlists: %w", err)
	}

	keys := make([]string, 0, len(simplePlaylists))
	for _, p := range simplePlaylists {
		keys = append(keys, fmt.Sprintf("playlist_%s", p.ID))
	}
	cachedPlaylists := make([]spotify.FullPlaylist, len(simplePlaylists))

	if err := s.kv.GetMany(ctx, keys, &cachedPlaylists); err != nil {
		return nil, fmt.Errorf("getting many playlists: %w", err)
	}

	playlists := make([]spotify.FullPlaylist, 0, len(simplePlaylists))
	for i, cached := range cachedPlaylists {
		// If it's populated and not outdated, use it
		if cached.SnapshotID != "" && !spotifyutil.SimplePlaylistHasChanged(cached.SimplePlaylist, simplePlaylists[i]) {
			playlists = append(playlists, cached)
		} else {
			reason := "not found"
			if cached.SnapshotID != "" {
				reason = "out of date"
			}
			slog.DebugContext(ctx, "getting populated playlist from Spotify", "playlist", simplePlaylists[i].Name, "reason", reason, "playlist_id", simplePlaylists[i].ID, "snapshot_id", simplePlaylists[i].SnapshotID)

			// Else fetch it from Spotify and store it
			playlist, err := s.getPlaylist(ctx, simplePlaylists[i].ID.String())
			if err != nil {
				return nil, fmt.Errorf("getting playlist %s: %w", simplePlaylists[i].ID, err)
			}
			playlists = append(playlists, playlist)
			if err := s.kv.Put(ctx, keys[i], playlist); err != nil {
				return nil, fmt.Errorf("storing playlist %s: %w", simplePlaylists[i].ID, err)
			}

			if err := s.kv.Put(ctx, keys[i], playlist); err != nil {
				return nil, fmt.Errorf("storing playlist %s (%s): %w", playlist.Name, playlist.ID, err)
			}
		}
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

		pgtr := paginator.New(paginator.InitialOffset(len(p.Tracks.Tracks)), paginator.Parallelism(10))
		itChan := make(chan indexAndTracks)
		g, ctx := errgroup.WithContext(ctx)
		g.Go(func() error {
			defer close(itChan)
			return pgtr.Run(ctx, func(i int, opts paginator.PageOpts, next paginator.NextFunc) (result *paginator.NextResult, err error) {
				page, err := s.spotify.GetPlaylistTracksOpt(spotify.ID(p.ID), spotifyOpts(opts), "")
				if err != nil {
					return nil, err
				}
				itChan <- indexAndTracks{i, page.Tracks}
				slog.Debug("listing playlist tracks", "playlist", p.Name, "counter", i, "offset", page.Offset, "total", page.Total)
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

	pgtr := paginator.New(paginator.Parallelism(10))
	ipChan := make(chan indexAndPlaylists)
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		defer close(ipChan)
		return pgtr.Run(ctx, func(i int, opts paginator.PageOpts, next paginator.NextFunc) (*paginator.NextResult, error) {
			page, err := s.spotify.GetPlaylistsForUserOpt(userID, spotifyOpts(opts))
			if err != nil {
				return nil, err
			}
			ipChan <- indexAndPlaylists{i, page.Playlists}
			slog.DebugContext(ctx, "listing playlists for user", "user", userID, "counter", i, "offset", page.Offset, "total", page.Total)
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
