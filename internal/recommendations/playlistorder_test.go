package recommendations

import (
	"context"
	"testing"

	"github.com/zmb3/spotify"
)

// Every playlist list in the UI is rendered in the order the API returns it, so
// the order is part of each endpoint's answer rather than the caller's problem.
func TestPlaylistsAreListedInDescendingNaturalOrder(t *testing.T) {
	want := []string{"Metal 600", "Metal 70", "Metal 9", "ambient"}

	t.Run("listing a user's playlists", func(t *testing.T) {
		svc := &service{spotify: &fakeSpotifyProvider{playlists: unorderedPlaylists()}}

		playlists, err := svc.ListPlaylistsForCurrentUser(context.Background())
		if err != nil {
			t.Fatalf("ListPlaylistsForCurrentUser() error = %v", err)
		}
		assertPlaylistOrder(t, want, playlists)
	})

	t.Run("the index summary the library window lists", func(t *testing.T) {
		index := &fakeTrackIndex{summary: IndexSummary{Playlists: unorderedPlaylists()}}
		svc := &service{spotify: &fakeSpotifyProvider{}, trackIndex: index, store: fakeKeyValueStore{}}

		summary, err := svc.GetIndexSummary(context.Background())
		if err != nil {
			t.Fatalf("GetIndexSummary() error = %v", err)
		}
		assertPlaylistOrder(t, want, summary.Playlists)
	})

	t.Run("the playlists a track belongs to", func(t *testing.T) {
		index := &fakeTrackIndex{lookup: unorderedPlaylists()}
		svc := &service{spotify: &fakeSpotifyProvider{}, trackIndex: index}

		playlists, err := svc.LookupTrackInLibrary(context.Background(), "track-1")
		if err != nil {
			t.Fatalf("LookupTrackInLibrary() error = %v", err)
		}
		assertPlaylistOrder(t, want, playlists)
	})
}

func unorderedPlaylists() []spotify.SimplePlaylist {
	return []spotify.SimplePlaylist{
		{Name: "Metal 70"},
		{Name: "ambient"},
		{Name: "Metal 600"},
		{Name: "Metal 9"},
	}
}

func assertPlaylistOrder(t *testing.T, want []string, playlists []spotify.SimplePlaylist) {
	t.Helper()
	got := make([]string, 0, len(playlists))
	for _, p := range playlists {
		got = append(got, p.Name)
	}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
