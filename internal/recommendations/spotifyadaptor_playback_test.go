package recommendations

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSpotifyAdaptorPlaybackState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/me/player" {
			t.Fatalf("request = %s %s, want GET /v1/me/player", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"timestamp": 1234,
			"progress_ms": 567,
			"is_playing": true,
			"item": {"id":"track-1","name":"Track one","duration_ms":240000},
			"device": {"id":"device-1","is_active":true,"is_restricted":false,"name":"Speaker","type":"Speaker","volume_percent":42},
			"shuffle_state": true,
			"repeat_state": "context",
			"context": {"type":"playlist","uri":"spotify:playlist:one"}
		}`))
	}))
	defer server.Close()

	adaptor := &SpotifyAdaptor{http: server.Client(), apiBaseURL: server.URL + "/v1/"}
	state, err := adaptor.PlaybackState(context.Background())
	if err != nil {
		t.Fatalf("PlaybackState() error = %v", err)
	}
	if state == nil || state.Item == nil || state.Item.ID.String() != "track-1" {
		t.Fatalf("PlaybackState() state = %#v", state)
	}
	if !state.Playing || state.Progress != 567 || state.Device.ID.String() != "device-1" {
		t.Fatalf("PlaybackState() state = %#v", state)
	}
}

func TestSpotifyAdaptorPlaybackStateHandlesNoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	adaptor := &SpotifyAdaptor{http: server.Client(), apiBaseURL: server.URL + "/"}
	state, err := adaptor.PlaybackState(context.Background())
	if err != nil {
		t.Fatalf("PlaybackState() error = %v", err)
	}
	if state != nil {
		t.Fatalf("PlaybackState() = %#v, want nil", state)
	}
}

func TestSpotifyAdaptorPlaybackQueue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/me/player/queue" {
			t.Fatalf("path = %s, want /v1/me/player/queue", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"currently_playing":{"id":"current","name":"Current"},
			"queue":[{"id":"next-1","name":"Next one"},{"id":"next-2","name":"Next two"}]
		}`))
	}))
	defer server.Close()

	adaptor := &SpotifyAdaptor{http: server.Client(), apiBaseURL: server.URL + "/v1/"}
	queue, err := adaptor.PlaybackQueue(context.Background())
	if err != nil {
		t.Fatalf("PlaybackQueue() error = %v", err)
	}
	if queue.CurrentlyPlaying == nil || queue.CurrentlyPlaying.ID.String() != "current" {
		t.Fatalf("PlaybackQueue() current = %#v", queue.CurrentlyPlaying)
	}
	if len(queue.Tracks) != 2 || queue.Tracks[1].ID.String() != "next-2" {
		t.Fatalf("PlaybackQueue() tracks = %#v", queue.Tracks)
	}
}

func TestSpotifyAdaptorReportsSpotifyErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"message":"slow down"}}`))
	}))
	defer server.Close()

	adaptor := &SpotifyAdaptor{http: server.Client(), apiBaseURL: server.URL + "/"}
	_, err := adaptor.PlaybackState(context.Background())
	if err == nil || !strings.Contains(err.Error(), "429") {
		t.Fatalf("PlaybackState() error = %v, want status 429", err)
	}
}
