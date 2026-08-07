package recommendations

import (
	"context"
	"testing"

	"github.com/zmb3/spotify"
)

func TestGetPlaybackMapsPlayerState(t *testing.T) {
	track := spotify.FullTrack{SimpleTrack: spotify.SimpleTrack{ID: spotify.ID("track-1"), Name: "Track"}}
	provider := &fakeSpotifyProvider{playbackState: &spotify.PlayerState{
		CurrentlyPlaying: spotify.CurrentlyPlaying{
			Timestamp: 1234,
			Progress:  567,
			Playing:   true,
			Item:      &track,
		},
		Device:       spotify.PlayerDevice{ID: spotify.ID("device-1"), Name: "Speaker", Active: true},
		ShuffleState: true,
		RepeatState:  "context",
	}}

	playback, err := (&service{spotify: provider}).GetPlayback(context.Background())
	if err != nil {
		t.Fatalf("GetPlayback() error = %v", err)
	}
	if !playback.Active || !playback.IsPlaying {
		t.Fatalf("GetPlayback() active/playing = %v/%v, want true/true", playback.Active, playback.IsPlaying)
	}
	if playback.Track == nil || playback.Track.ID != track.ID {
		t.Fatalf("GetPlayback() track = %#v, want %s", playback.Track, track.ID)
	}
	if playback.Device == nil || playback.Device.ID != spotify.ID("device-1") {
		t.Fatalf("GetPlayback() device = %#v", playback.Device)
	}
	if playback.ProgressMs != 567 || playback.Timestamp != 1234 {
		t.Fatalf("GetPlayback() progress/timestamp = %d/%d", playback.ProgressMs, playback.Timestamp)
	}
}

func TestGetPlaybackWithoutActiveDevice(t *testing.T) {
	playback, err := (&service{spotify: &fakeSpotifyProvider{}}).GetPlayback(context.Background())
	if err != nil {
		t.Fatalf("GetPlayback() error = %v", err)
	}
	if playback.Active || playback.Track != nil || playback.Device != nil {
		t.Fatalf("GetPlayback() = %#v, want inactive playback", playback)
	}
}

func TestSkipPlaybackQueueValidatesAndAdvances(t *testing.T) {
	current := fullTrack("current")
	provider := &fakeSpotifyProvider{queue: PlaybackQueue{
		CurrentlyPlaying: &current,
		Tracks: []spotify.FullTrack{
			fullTrack("one"),
			fullTrack("two"),
			fullTrack("three"),
		},
	}}
	svc := &service{spotify: provider}

	err := svc.SkipPlaybackQueue(context.Background(), QueueSkipRequest{
		Position:               1,
		ExpectedTrackID:        "two",
		ExpectedCurrentTrackID: "current",
	})
	if err != nil {
		t.Fatalf("SkipPlaybackQueue() error = %v", err)
	}
	if provider.nextCalls != 2 {
		t.Fatalf("SkipPlaybackQueue() Next calls = %d, want 2", provider.nextCalls)
	}
}

func TestSkipPlaybackQueueRejectsStaleQueue(t *testing.T) {
	provider := &fakeSpotifyProvider{queue: PlaybackQueue{Tracks: []spotify.FullTrack{fullTrack("one")}}}
	err := (&service{spotify: provider}).SkipPlaybackQueue(context.Background(), QueueSkipRequest{
		Position:        0,
		ExpectedTrackID: "different",
	})
	if err != ErrQueueChanged {
		t.Fatalf("SkipPlaybackQueue() error = %v, want ErrQueueChanged", err)
	}
	if provider.nextCalls != 0 {
		t.Fatalf("SkipPlaybackQueue() Next calls = %d, want 0", provider.nextCalls)
	}
}

func TestGetIndexSummaryDoesNotSynchronize(t *testing.T) {
	provider := &fakeSpotifyProvider{user: spotify.User{ID: "user-1"}}
	index := &fakeTrackIndex{summary: IndexSummary{PlaylistCount: 2, UniqueTrackCount: 7}}
	svc := &service{spotify: provider, trackIndex: index, store: fakeKeyValueStore{}}

	summary, err := svc.GetIndexSummary(context.Background())
	if err != nil {
		t.Fatalf("GetIndexSummary() error = %v", err)
	}
	if summary.PlaylistCount != 2 || summary.UniqueTrackCount != 7 {
		t.Fatalf("GetIndexSummary() = %#v", summary)
	}
	if provider.listPlaylistCalls != 0 || index.diffCalls != 0 || index.syncCalls != 0 {
		t.Fatalf("GetIndexSummary() synchronized: list=%d diff=%d sync=%d", provider.listPlaylistCalls, index.diffCalls, index.syncCalls)
	}
}

func fullTrack(id string) spotify.FullTrack {
	return spotify.FullTrack{SimpleTrack: spotify.SimpleTrack{ID: spotify.ID(id), Name: id}}
}

type fakeSpotifyProvider struct {
	user              spotify.User
	playbackState     *spotify.PlayerState
	queue             PlaybackQueue
	nextCalls         int
	listPlaylistCalls int
}

func (f *fakeSpotifyProvider) ListPlaylists(context.Context, string) ([]spotify.SimplePlaylist, error) {
	f.listPlaylistCalls++
	return nil, nil
}
func (f *fakeSpotifyProvider) GetPlaylist(context.Context, string) (spotify.FullPlaylist, error) {
	return spotify.FullPlaylist{}, nil
}
func (f *fakeSpotifyProvider) PopulatePlaylists(context.Context, []spotify.SimplePlaylist) ([]spotify.FullPlaylist, error) {
	return nil, nil
}
func (f *fakeSpotifyProvider) CreatePlaylist(context.Context, string, string, []string) (spotify.FullPlaylist, error) {
	return spotify.FullPlaylist{}, nil
}
func (f *fakeSpotifyProvider) SetPlaylistTracks(context.Context, string, []string) (spotify.FullPlaylist, error) {
	return spotify.FullPlaylist{}, nil
}
func (f *fakeSpotifyProvider) TruncatePlaylist(context.Context, string, string) error { return nil }
func (f *fakeSpotifyProvider) CurrentUser(context.Context) (spotify.User, error)      { return f.user, nil }
func (f *fakeSpotifyProvider) CurrentTrack(context.Context) (spotify.FullTrack, bool, error) {
	return spotify.FullTrack{}, false, nil
}
func (f *fakeSpotifyProvider) GetAlbum(context.Context, string) (spotify.FullAlbum, error) {
	return spotify.FullAlbum{}, nil
}
func (f *fakeSpotifyProvider) GetAlbums(context.Context, []string) ([]spotify.FullAlbum, error) {
	return nil, nil
}
func (f *fakeSpotifyProvider) ListArtistAlbums(context.Context, string) ([]spotify.SimpleAlbum, error) {
	return nil, nil
}
func (f *fakeSpotifyProvider) GetTrack(context.Context, string) (spotify.FullTrack, error) {
	return spotify.FullTrack{}, nil
}
func (f *fakeSpotifyProvider) PlaybackState(context.Context) (*spotify.PlayerState, error) {
	return f.playbackState, nil
}
func (f *fakeSpotifyProvider) PlaybackQueue(context.Context) (PlaybackQueue, error) {
	return f.queue, nil
}
func (f *fakeSpotifyProvider) RecentlyPlayed(context.Context, int) ([]spotify.RecentlyPlayedItem, error) {
	return nil, nil
}
func (f *fakeSpotifyProvider) Play(context.Context, string) error { return nil }
func (f *fakeSpotifyProvider) Pause(context.Context) error        { return nil }
func (f *fakeSpotifyProvider) Next(context.Context) error {
	f.nextCalls++
	return nil
}
func (f *fakeSpotifyProvider) Previous(context.Context) error  { return nil }
func (f *fakeSpotifyProvider) Seek(context.Context, int) error { return nil }

type fakeTrackIndex struct {
	summary   IndexSummary
	diffCalls int
	syncCalls int
}

func (f *fakeTrackIndex) Has(context.Context, string, spotify.SimpleTrack) (bool, error) {
	return false, nil
}
func (f *fakeTrackIndex) Lookup(context.Context, string, spotify.SimpleTrack) ([]spotify.SimplePlaylist, error) {
	return nil, nil
}
func (f *fakeTrackIndex) Diff(context.Context, string, []spotify.SimplePlaylist) ([]spotify.SimplePlaylist, []spotify.SimplePlaylist, []spotify.SimplePlaylist, error) {
	f.diffCalls++
	return nil, nil, nil, nil
}
func (f *fakeTrackIndex) Sync(context.Context, string, []spotify.FullPlaylist, []spotify.FullPlaylist, []spotify.FullPlaylist) error {
	f.syncCalls++
	return nil
}
func (f *fakeTrackIndex) CountTracksByArtist(context.Context, string, string) (int, error) {
	return 0, nil
}
func (f *fakeTrackIndex) Summarize(context.Context, string) (IndexSummary, error) {
	return f.summary, nil
}

type fakeKeyValueStore struct{}

func (fakeKeyValueStore) Get(context.Context, string, any) (bool, error) { return false, nil }
func (fakeKeyValueStore) GetMany(context.Context, []string, any) error   { return nil }
func (fakeKeyValueStore) Put(context.Context, string, any) error         { return nil }
