package recommendations

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

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
	PlaylistCount    int                      `json:"playlist_count"`
	UniqueTrackCount int                      `json:"unique_track_count"`
	Playlists        []spotify.SimplePlaylist `json:"playlists"`
	LastSyncedAt     *time.Time               `json:"last_synced_at,omitempty"`
}

func indexSyncKey(userID string) string {
	return fmt.Sprintf("track-index:last-synced:%s", userID)
}

func (s *service) getIndexSummary(ctx context.Context, userID string) (IndexSummary, error) {
	summary, err := s.trackIndex.Summarize(ctx, userID)
	if err != nil {
		return IndexSummary{}, fmt.Errorf("summarizing track index: %w", err)
	}

	var syncedAt time.Time
	if exists, err := s.store.Get(ctx, indexSyncKey(userID), &syncedAt); err != nil {
		return IndexSummary{}, fmt.Errorf("getting track index sync time: %w", err)
	} else if exists {
		summary.LastSyncedAt = &syncedAt
	}
	if summary.Playlists == nil {
		summary.Playlists = []spotify.SimplePlaylist{}
	}
	return summary, nil
}

func TrackKey(tt spotify.SimpleTrack) string {
	artistNames := make([]string, 0, len(tt.Artists))
	for _, a := range tt.Artists {
		artistNames = append(artistNames, a.Name)
	}
	sort.Strings(artistNames)

	return fmt.Sprintf("%s - %s", tt.Name, strings.Join(artistNames, ", "))
}
