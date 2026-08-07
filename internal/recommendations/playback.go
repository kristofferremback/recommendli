package recommendations

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zmb3/spotify"
)

var (
	ErrNoActivePlayback = errors.New("no active Spotify playback")
	ErrQueueChanged     = errors.New("Spotify queue changed; refresh it and try again")
)

// Playback is the lightweight state polled by the remote player. ProgressMs is
// the position at Timestamp, allowing the UI to animate progress locally
// between requests.
type Playback struct {
	Active       bool                     `json:"active"`
	IsPlaying    bool                     `json:"is_playing"`
	ProgressMs   int                      `json:"progress_ms"`
	Timestamp    int64                    `json:"timestamp"`
	Track        *spotify.FullTrack       `json:"track,omitempty"`
	Device       *spotify.PlayerDevice    `json:"device,omitempty"`
	Context      *spotify.PlaybackContext `json:"context,omitempty"`
	ShuffleState bool                     `json:"shuffle_state"`
	RepeatState  string                   `json:"repeat_state"`
}

type PlaybackQueue struct {
	CurrentlyPlaying *spotify.FullTrack  `json:"currently_playing,omitempty"`
	Tracks           []spotify.FullTrack `json:"tracks"`
}

type QueueSkipRequest struct {
	Position               int    `json:"position"`
	ExpectedTrackID        string `json:"expected_track_id"`
	ExpectedCurrentTrackID string `json:"expected_current_track_id"`
}

func (s *service) GetPlayback(ctx context.Context) (Playback, error) {
	state, err := s.spotify.PlaybackState(ctx)
	if err != nil {
		return Playback{}, fmt.Errorf("getting playback state: %w", err)
	}
	if state == nil || state.Item == nil {
		return Playback{Active: false}, nil
	}

	return Playback{
		Active:       true,
		IsPlaying:    state.Playing,
		ProgressMs:   state.Progress,
		Timestamp:    state.Timestamp,
		Track:        state.Item,
		Device:       &state.Device,
		Context:      &state.PlaybackContext,
		ShuffleState: state.ShuffleState,
		RepeatState:  state.RepeatState,
	}, nil
}

func (s *service) GetPlaybackQueue(ctx context.Context) (PlaybackQueue, error) {
	queue, err := s.spotify.PlaybackQueue(ctx)
	if err != nil {
		return PlaybackQueue{}, fmt.Errorf("getting playback queue: %w", err)
	}
	if queue.Tracks == nil {
		queue.Tracks = []spotify.FullTrack{}
	}
	return queue, nil
}

func (s *service) GetPlaybackHistory(ctx context.Context, limit int) ([]spotify.RecentlyPlayedItem, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	items, err := s.spotify.RecentlyPlayed(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("getting recently played tracks: %w", err)
	}
	if items == nil {
		items = []spotify.RecentlyPlayedItem{}
	}
	return items, nil
}

func (s *service) Play(ctx context.Context, trackID string) error {
	if err := s.spotify.Play(ctx, trackID); err != nil {
		return fmt.Errorf("starting playback: %w", err)
	}
	return nil
}

func (s *service) Pause(ctx context.Context) error {
	if err := s.spotify.Pause(ctx); err != nil {
		return fmt.Errorf("pausing playback: %w", err)
	}
	return nil
}

func (s *service) Next(ctx context.Context) error {
	if err := s.spotify.Next(ctx); err != nil {
		return fmt.Errorf("skipping to next track: %w", err)
	}
	return nil
}

func (s *service) Previous(ctx context.Context) error {
	if err := s.spotify.Previous(ctx); err != nil {
		return fmt.Errorf("skipping to previous track: %w", err)
	}
	return nil
}

func (s *service) Seek(ctx context.Context, positionMs int) error {
	if positionMs < 0 {
		return errors.New("position_ms cannot be negative")
	}
	if err := s.spotify.Seek(ctx, positionMs); err != nil {
		return fmt.Errorf("seeking playback: %w", err)
	}
	return nil
}

// SkipPlaybackQueue advances through Spotify's queue. The expected IDs protect
// against acting on a stale UI after Spotify changes the queue between polling
// and the command reaching the server.
func (s *service) SkipPlaybackQueue(ctx context.Context, req QueueSkipRequest) error {
	if req.Position < 0 {
		return errors.New("position cannot be negative")
	}

	queue, err := s.GetPlaybackQueue(ctx)
	if err != nil {
		return err
	}
	if req.Position >= len(queue.Tracks) {
		return errors.New("position is outside the current queue")
	}
	if req.ExpectedTrackID == "" || queue.Tracks[req.Position].ID.String() != req.ExpectedTrackID {
		return ErrQueueChanged
	}
	if req.ExpectedCurrentTrackID != "" {
		if queue.CurrentlyPlaying == nil || queue.CurrentlyPlaying.ID.String() != req.ExpectedCurrentTrackID {
			return ErrQueueChanged
		}
	}

	for i := 0; i <= req.Position; i++ {
		if err := s.Next(ctx); err != nil {
			return err
		}
		// Give Spotify a short window to apply each command before sending the
		// next one. Queue jumps are user initiated and queues are short.
		if i < req.Position {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
			}
		}
	}
	return nil
}
