package recommendations

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zmb3/spotify"
)

type TrackPlaylistIndex struct {
	Playlists          map[string]spotify.SimplePlaylist
	TrackPlaylistIndex map[string][]string
	Tracks             map[string]spotify.SimpleTrack
}

type IndexSummary struct {
	Playlists    int
	UniqueTracks int
}

func NewTrackPlaylistIndexFromFullPlaylists(playlists []spotify.FullPlaylist) *TrackPlaylistIndex {
	index := &TrackPlaylistIndex{
		Playlists:          make(map[string]spotify.SimplePlaylist),
		TrackPlaylistIndex: make(map[string][]string),
		Tracks:             make(map[string]spotify.SimpleTrack),
	}
	for _, p := range playlists {
		for _, t := range p.Tracks.Tracks {
			index.Add(t.Track, p.SimplePlaylist)
		}
	}
	return index
}

func (t *TrackPlaylistIndex) Key(tt spotify.FullTrack) string {
	artistNames := make([]string, 0, len(tt.Artists))
	for _, a := range tt.Artists {
		artistNames = append(artistNames, a.Name)
	}
	sort.Strings(artistNames)

	return fmt.Sprintf("%s - %s", tt.Name, strings.Join(artistNames, ", "))
}

func (t *TrackPlaylistIndex) Add(tt spotify.FullTrack, p spotify.SimplePlaylist) {
	t.TrackPlaylistIndex[t.Key(tt)] = append(t.TrackPlaylistIndex[t.Key(tt)], p.ID.String())
	t.Tracks[t.Key(tt)] = tt.SimpleTrack
	if _, exists := t.Playlists[p.ID.String()]; !exists {
		t.Playlists[p.ID.String()] = p
	}
}

func (t *TrackPlaylistIndex) Lookup(tt spotify.FullTrack) ([]spotify.SimplePlaylist, bool) {
	playlistIDs, ok := t.TrackPlaylistIndex[t.Key(tt)]
	playlists := make([]spotify.SimplePlaylist, 0, len(playlistIDs))
	for _, id := range playlistIDs {
		playlists = append(playlists, t.Playlists[id])
	}
	return playlists, ok
}

func (t *TrackPlaylistIndex) Has(tt spotify.FullTrack) bool {
	_, ok := t.TrackPlaylistIndex[t.Key(tt)]
	return ok
}

func (t *TrackPlaylistIndex) MatchesSimplePlaylists(simplePlaylists []spotify.SimplePlaylist) bool {
	if len(t.Playlists) != len(simplePlaylists) {
		return false
	}
	snapshotIDs := make(map[string]string)
	for _, p := range simplePlaylists {
		snapshotIDs[p.ID.String()] = p.SnapshotID
	}
	for _, p := range t.Playlists {
		if snapshotID, found := snapshotIDs[p.ID.String()]; !found || p.SnapshotID != snapshotID {
			return false
		}
	}
	return true
}
