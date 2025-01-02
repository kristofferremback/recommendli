package recommendations

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/zmb3/spotify"
)

type TrackIndex interface {
	Has(ctx context.Context, userID string, track spotify.SimpleTrack) (bool, error)
	Lookup(ctx context.Context, userID string, track spotify.SimpleTrack) ([]spotify.SimplePlaylist, error)
	Diff(ctx context.Context, userID string, playlists []spotify.SimplePlaylist) (added, changed, removed []spotify.SimplePlaylist, err error)
	Sync(ctx context.Context, userID string, added, changed, removed []spotify.FullPlaylist) error
	CountTracksByArtist(ctx context.Context, userID string, artistName string) (int, error)
	Summarize(ctx context.Context, userID string) (IndexSummary, error)
}

type IndexSummary struct {
	PlaylistCount    int
	UniqueTrackCount int
	Playlists        []spotify.SimplePlaylist
}

func TrackKey(tt spotify.SimpleTrack) string {
	artistNames := make([]string, 0, len(tt.Artists))
	for _, a := range tt.Artists {
		artistNames = append(artistNames, a.Name)
	}
	sort.Strings(artistNames)

	return fmt.Sprintf("%s - %s", tt.Name, strings.Join(artistNames, ", "))
}
