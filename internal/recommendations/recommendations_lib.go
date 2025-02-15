package recommendations

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/kristofferostlund/recommendli/pkg/paginator"
	"github.com/kristofferostlund/recommendli/pkg/slogutil"
	"github.com/zmb3/spotify"
)

func (s *service) getPlaylistsAndSyncIndex(ctx context.Context, userID string) ([]spotify.SimplePlaylist, error) {
	// TODO: Make it possible to check if the index has been synced very recently
	ctx = slogutil.WithAttrs(ctx, slog.String("user", userID))

	key := fmt.Sprintf("getPlaylistsAndSyncIndex:%s", userID)

	playlists, err := s.sfSyncIndex(ctx, key, func(ctx context.Context) ([]spotify.SimplePlaylist, error) {
		playlists, err := s.spotify.ListPlaylists(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("listing user playlists generating discovery playlist: %w", err)
		}

		prefs, err := s.userPreferences.GetPreferences(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("getting user preferences: %w", err)
		}

		library := filterSimplePlaylist(playlists, func(p spotify.SimplePlaylist) bool {
			return prefs.IsLibraryPlaylistName(p.Name)
		})

		slog.DebugContext(ctx, "ensuring track index is synced")

		added, changed, removed, err := s.trackIndex.Diff(ctx, userID, library)
		if err != nil {
			return nil, fmt.Errorf("checking if track index needs sync: %w", err)
		}

		if len(added)+len(changed)+len(removed) == 0 {
			slog.DebugContext(ctx, "track index already up to date")
			return playlists, nil
		}

		slog.DebugContext(ctx, "populating tracks for track index", "added", len(added), "changed", len(changed), "removed", len(removed))

		addedPlaylists, err := s.spotify.PopulatePlaylists(ctx, added)
		if err != nil {
			return nil, fmt.Errorf("populating added playlists: %w", err)
		}
		changedPlaylists, err := s.spotify.PopulatePlaylists(ctx, changed)
		if err != nil {
			return nil, fmt.Errorf("populating changed playlists: %w", err)
		}
		removedPlaylists, err := s.spotify.PopulatePlaylists(ctx, removed)
		if err != nil {
			return nil, fmt.Errorf("populating removed playlists: %w", err)
		}

		slog.DebugContext(ctx, "syncing track index")

		if err := s.trackIndex.Sync(ctx, userID, addedPlaylists, changedPlaylists, removedPlaylists); err != nil {
			return nil, fmt.Errorf("syncing track index: %w", err)
		}

		slog.DebugContext(ctx, "track index successfully synced")

		return playlists, nil
	})
	if err != nil {
		return nil, fmt.Errorf("getting playlists and syncing index: %w", err)
	}

	return playlists, nil
}

func (s *service) albumForTrack(ctx context.Context, track spotify.FullTrack) (spotify.FullAlbum, error) {
	slog.DebugContext(ctx, "getting album for track", "track", stringifyTrack(track.SimpleTrack), "album", track.Album.Name, "album_id", track.Album.ID.String())
	album, err := s.spotify.GetAlbum(ctx, track.Album.ID.String())
	if err != nil {
		return spotify.FullAlbum{}, fmt.Errorf("getting album for track %s: %w", track.ID, err)
	}
	if album.AlbumType == "album" {
		return album, nil
	}

	artistIndex := make(map[string]int)
	for i, artist := range track.Artists {
		artistIndex[artist.ID.String()] = i
	}

	simpleAlbums := make([]spotify.SimpleAlbum, 0)
	for _, artist := range track.Artists {
		sa, err := s.spotify.ListArtistAlbums(ctx, artist.ID.String())
		if err != nil {
			return spotify.FullAlbum{}, err
		}
		simpleAlbums = append(simpleAlbums, sa...)
	}
	if len(simpleAlbums) == 0 || (len(simpleAlbums) == 1 && simpleAlbums[0].ID == album.ID) {
		return album, nil
	}

	albumIDs := make([]string, 0)
	for _, a := range simpleAlbums {
		albumIDs = append(albumIDs, a.ID.String())
	}
	albums, err := s.spotify.GetAlbums(ctx, albumIDs)
	if err != nil {
		return spotify.FullAlbum{}, err
	}
	sort.Slice(albums, func(i, j int) bool {
		// the first named artist is really the most important
		if albums[i].Artists[0].ID != albums[j].Artists[0].ID {
			return artistIndex[albums[i].Artists[0].ID.String()] < artistIndex[albums[j].Artists[0].ID.String()]
		}
		// albums are preferred over other releases
		if albums[i].AlbumType == "album" && albums[j].AlbumType != "album" {
			return true
		}
		// otherwise we prefer more recent releases
		if !albums[i].ReleaseDateTime().Equal(albums[j].ReleaseDateTime()) {
			return albums[i].ReleaseDateTime().Before(albums[j].ReleaseDateTime())
		}
		// lastly we prefer the largest albums
		return albums[i].Tracks.Total < albums[j].Tracks.Total
	})

	for _, a := range albums {
		for _, t := range a.Tracks.Tracks {
			// Using this, we don't consider a track with differing artists to be the same track.
			// Example:
			// - https://open.spotify.com/track/4EllS6NXwCOggtnnBqByNd?si=c67438dea8264191 (Warrior by Atreyu, Travis Barker)
			// - https://open.spotify.com/track/0KBRMpZVUTxrU8blRUBfm3?si=567abdce7f174e88 (Warrior by Atreyo, Zero 9:36, Travis Barker)
			// Only checking by name the above two tracks match, however the second track is in fact a different track.
			if stringifyTrack(t) == stringifyTrack(track.SimpleTrack) {
				return a, nil
			}
		}
	}

	return album, nil
}

func (s *service) trackAndAlbum(ctx context.Context, track spotify.FullTrack) (spotify.FullTrack, spotify.FullAlbum, error) {
	album, err := s.albumForTrack(ctx, track)
	if err != nil {
		return spotify.FullTrack{}, spotify.FullAlbum{}, err
	}
	if album.ID == track.Album.ID {
		return track, album, nil
	}
	for _, t := range album.Tracks.Tracks {
		if t.Name == track.Name {
			tt, err := s.spotify.GetTrack(ctx, t.ID.String())
			if err != nil {
				return spotify.FullTrack{}, spotify.FullAlbum{}, err
			}
			return tt, album, nil
		}
	}
	return track, album, nil
}

func (s *service) scoreTracks(ctx context.Context, userID string, tracks []spotify.FullTrack) ([]score, error) {
	type indexAndTrack struct {
		index  int
		scores []score
	}
	pgtr := paginator.New(
		paginator.Parallelism(10),
		paginator.PageSize(1),
		paginator.InitialTotalCount(len(tracks)),
	)
	trackChan := make(chan indexAndTrack)

	errC := make(chan error)
	done := make(chan struct{})

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		defer close(trackChan)
		if err := pgtr.Run(ctx, func(i int, opts paginator.PageOpts, next paginator.NextFunc) (result *paginator.NextResult, err error) {
			scores := make([]score, 0)
			from, to := opts.Offset, opts.Offset+opts.Limit
			for _, t := range tracks[from:to] {
				if t.ID.String() == "" {
					slog.DebugContext(ctx, "skipping track with empty ID", "track", stringifyTrack(t.SimpleTrack))
					continue
				}
				track, album, err := s.trackAndAlbum(ctx, t)
				if err != nil {
					return nil, err
				}
				artistRelevace := 0
				for _, a := range track.Artists {
					ar, err := s.trackIndex.CountTracksByArtist(ctx, userID, a.Name)
					if err != nil {
						return nil, fmt.Errorf("counting tracks by artist %s: %w", a.Name, err)
					}
					artistRelevace += ar
				}
				scores = append(scores, score{track: track, album: album, artistRelevace: artistRelevace})
			}
			slog.DebugContext(ctx, "getting most relevant tracks", "total count", len(tracks), "batch size", to-from, "from", from, "to", to)
			trackChan <- indexAndTrack{i, scores}
			return next(len(tracks)), nil
		}); err != nil {
			errC <- err
			return
		}

		done <- struct{}{}
	}()

	indexedScores := make([]indexAndTrack, 0)
loop:
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err := <-errC:
			return nil, err
		case it := <-trackChan:
			indexedScores = append(indexedScores, it)
		case <-done:
			break loop
		}
	}

	mostRelevant := make([]score, 0)
	for _, indexed := range indexedScores {
		mostRelevant = append(mostRelevant, indexed.scores...)
	}

	return mostRelevant, nil
}

func (s *service) upsertPlaylistByName(ctx context.Context, existingPlaylists []spotify.SimplePlaylist, userID, playlistName string, trackIDs []string) (spotify.FullPlaylist, error) {
	for _, p := range existingPlaylists {
		if p.Name == playlistName {
			if err := s.spotify.TruncatePlaylist(ctx, p.ID.String(), p.SnapshotID); err != nil {
				return spotify.FullPlaylist{}, fmt.Errorf("truncating playlist %s: %w", p.ID, err)
			}
			return s.spotify.SetPlaylistTracks(ctx, p.ID.String(), trackIDs)
		}
	}
	return s.spotify.CreatePlaylist(ctx, userID, playlistName, trackIDs)
}

func dummyPlaylistFor(name string, tracks []spotify.FullTrack) spotify.FullPlaylist {
	pl := spotify.FullPlaylist{
		SimplePlaylist: spotify.SimplePlaylist{
			Name: fmt.Sprintf("dummy: %s", name),
			Tracks: spotify.PlaylistTracks{
				Total: uint(len(tracks)),
			},
		},
	}
	pl.Tracks.Total = len(tracks)
	for _, t := range tracks {
		pl.Tracks.Tracks = append(pl.Tracks.Tracks, spotify.PlaylistTrack{Track: t})
	}
	return pl
}

func filterSimplePlaylist(playlists []spotify.SimplePlaylist, pred func(p spotify.SimplePlaylist) bool) []spotify.SimplePlaylist {
	filtered := make([]spotify.SimplePlaylist, 0)
	for _, p := range playlists {
		if pred(p) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func stringsContain(ss []string, v string) bool {
	for _, s := range ss {
		if s == v {
			return true
		}
	}
	return false
}

func tracksFor(playlists []spotify.FullPlaylist) []spotify.FullTrack {
	tracks := make([]spotify.FullTrack, 0)
	for _, p := range playlists {
		tracks = append(tracks, tracksOf(p)...)
	}
	return tracks
}

func tracksOf(p spotify.FullPlaylist) []spotify.FullTrack {
	tracks := make([]spotify.FullTrack, 0)
	for _, t := range p.Tracks.Tracks {
		tracks = append(tracks, t.Track)
	}
	return tracks
}

func trackIDsOf(tracks []spotify.FullTrack) []string {
	trackIDs := make([]string, 0)
	for _, t := range tracks {
		trackIDs = append(trackIDs, t.ID.String())
	}
	return trackIDs
}

func stringifyTrack(t spotify.SimpleTrack) string {
	artistNames := make([]string, 0, len(t.Artists))
	for _, a := range t.Artists {
		artistNames = append(artistNames, a.Name)
	}
	sort.Strings(artistNames)

	name, artists := t.Name, strings.Join(artistNames, ", ")
	if name == "" {
		name = "<Unknown track>"
	}
	if artists == "" {
		artists = "<Unknown artist>"
	}

	return fmt.Sprintf("%s - %s", name, artists)
}

func printableTrack(t spotify.SimpleTrack) string {
	return fmt.Sprintf("%s - %s", t.URI, stringifyTrack(t))
}

func printableTracks(tracks []spotify.FullTrack) []string {
	printable := make([]string, 0)
	for _, t := range tracks {
		printable = append(printable, printableTrack(t.SimpleTrack))
	}
	return printable
}

func uniqueTracks(tracks []spotify.FullTrack) []spotify.FullTrack {
	seen := make(map[string]struct{})
	unique := make([]spotify.FullTrack, 0)
	for _, t := range tracks {
		if _, isSeen := seen[stringifyTrack(t.SimpleTrack)]; !isSeen {
			seen[stringifyTrack(t.SimpleTrack)] = struct{}{}
			unique = append(unique, t)
		}
	}
	return unique
}
