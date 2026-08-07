package recommendations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/kristofferostlund/recommendli/pkg/ctxhelper"
	"github.com/zmb3/spotify"
)

const spotifyAPIBaseURL = "https://api.spotify.com/v1/"

func (s *SpotifyAdaptor) PlaybackState(ctx context.Context) (*spotify.PlayerState, error) {
	var state spotify.PlayerState
	found, err := s.getSpotifyAPI(ctx, "me/player", &state)
	if err != nil {
		return nil, fmt.Errorf("getting playback state: %w", err)
	}
	if !found {
		return nil, nil
	}
	return &state, nil
}

func (s *SpotifyAdaptor) PlaybackQueue(ctx context.Context) (PlaybackQueue, error) {
	var response struct {
		CurrentlyPlaying *spotify.FullTrack  `json:"currently_playing"`
		Queue            []spotify.FullTrack `json:"queue"`
	}
	found, err := s.getSpotifyAPI(ctx, "me/player/queue", &response)
	if err != nil {
		return PlaybackQueue{}, fmt.Errorf("getting playback queue: %w", err)
	}
	if !found {
		return PlaybackQueue{Tracks: []spotify.FullTrack{}}, nil
	}
	return PlaybackQueue{
		CurrentlyPlaying: response.CurrentlyPlaying,
		Tracks:           response.Queue,
	}, nil
}

func (s *SpotifyAdaptor) RecentlyPlayed(ctx context.Context, limit int) ([]spotify.RecentlyPlayedItem, error) {
	if err := ctxhelper.Closed(ctx); err != nil {
		return nil, fmt.Errorf("getting recently played tracks: %w", err)
	}
	items, err := s.spotify.PlayerRecentlyPlayedOpt(&spotify.RecentlyPlayedOptions{Limit: limit})
	if err != nil {
		return nil, fmt.Errorf("getting recently played tracks: %w", err)
	}
	return items, nil
}

func (s *SpotifyAdaptor) Play(ctx context.Context, trackID string) error {
	if err := ctxhelper.Closed(ctx); err != nil {
		return fmt.Errorf("starting playback: %w", err)
	}
	if trackID == "" {
		return s.spotify.Play()
	}
	return s.spotify.PlayOpt(&spotify.PlayOptions{
		URIs: []spotify.URI{spotify.URI("spotify:track:" + trackID)},
	})
}

func (s *SpotifyAdaptor) Pause(ctx context.Context) error {
	if err := ctxhelper.Closed(ctx); err != nil {
		return fmt.Errorf("pausing playback: %w", err)
	}
	return s.spotify.Pause()
}

func (s *SpotifyAdaptor) Next(ctx context.Context) error {
	if err := ctxhelper.Closed(ctx); err != nil {
		return fmt.Errorf("skipping to next track: %w", err)
	}
	return s.spotify.Next()
}

func (s *SpotifyAdaptor) Previous(ctx context.Context) error {
	if err := ctxhelper.Closed(ctx); err != nil {
		return fmt.Errorf("skipping to previous track: %w", err)
	}
	return s.spotify.Previous()
}

func (s *SpotifyAdaptor) Seek(ctx context.Context, positionMs int) error {
	if err := ctxhelper.Closed(ctx); err != nil {
		return fmt.Errorf("seeking playback: %w", err)
	}
	return s.spotify.Seek(positionMs)
}

// getSpotifyAPI is used for endpoints missing from spotify v1.3.0 and for
// playback-state 204 responses, which that version does not model cleanly.
func (s *SpotifyAdaptor) getSpotifyAPI(ctx context.Context, path string, out any) (bool, error) {
	if err := ctxhelper.Closed(ctx); err != nil {
		return false, err
	}
	if s.http == nil {
		return false, errMissingHTTPClient
	}
	baseURL := s.apiBaseURL
	if baseURL == "" {
		baseURL = spotifyAPIBaseURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path, nil)
	if err != nil {
		return false, err
	}
	res, err := s.http.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNoContent {
		return false, nil
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return false, fmt.Errorf("Spotify API returned %s: %s", res.Status, string(body))
	}
	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return false, err
	}
	return true, nil
}

var errMissingHTTPClient = errors.New("authenticated Spotify HTTP client is unavailable")
