package recommendations

import (
	"context"
	"fmt"

	"github.com/kristofferostlund/recommendli/pkg/ctxhelper"
	"github.com/kristofferostlund/recommendli/pkg/logging"
	"github.com/kristofferostlund/recommendli/pkg/paginator"
	"github.com/zmb3/spotify"
)

type SpotifyAdaptor struct {
	spotify spotify.Client
	log     logging.Logger
	kv      KeyValueStore
}

type SpotifyAdaptorFactory struct {
	log   logging.Logger
	store KeyValueStore
}

func NewSpotifyProviderFactory(log logging.Logger, store KeyValueStore) *SpotifyAdaptorFactory {
	return &SpotifyAdaptorFactory{log: log, store: store}
}

func (f *SpotifyAdaptorFactory) New(spotifyClient spotify.Client) *SpotifyAdaptor {
	return &SpotifyAdaptor{spotify: spotifyClient, log: f.log, kv: f.store}
}

func (s *SpotifyAdaptor) CurrentUser(ctx context.Context) (spotify.User, error) {
	if err := ctxhelper.Closed(ctx); err != nil {
		return spotify.User{}, fmt.Errorf("getting current user: %w", err)
	}
	usr, err := s.spotify.CurrentUser()
	if err != nil {
		return spotify.User{}, fmt.Errorf("getting current user: %w", err)
	}
	return usr.User, nil
}

func (s *SpotifyAdaptor) CurrentTrack(ctx context.Context) (spotify.FullTrack, bool, error) {
	if err := ctxhelper.Closed(ctx); err != nil {
		return spotify.FullTrack{}, false, fmt.Errorf("getting currently playing track: %w", err)
	}
	p, err := s.spotify.PlayerCurrentlyPlaying()
	if err != nil {
		return spotify.FullTrack{}, false, fmt.Errorf("getting currently playing track: %w", err)
	}
	if !p.Playing {
		return spotify.FullTrack{}, false, nil
	}
	return *p.Item, true, nil
}

func (s *SpotifyAdaptor) GetTrack(ctx context.Context, trackID string) (spotify.FullTrack, error) {
	if err := ctxhelper.Closed(ctx); err != nil {
		return spotify.FullTrack{}, fmt.Errorf("getting track %s: %w", trackID, err)
	}
	storeKey := fmt.Sprintf("track_%s", trackID)
	var stored spotify.FullTrack
	if exists, err := s.kv.Get(ctx, storeKey, &stored); err == nil && exists {
		return stored, nil
	} else if err != nil {
		return spotify.FullTrack{}, fmt.Errorf("getting track %s from store: %w", trackID, err)
	}

	track, err := s.spotify.GetTrack(spotify.ID(trackID))
	if err != nil {
		return spotify.FullTrack{}, fmt.Errorf("getting track %s: %w", trackID, err)
	}
	if track == nil {
		return spotify.FullTrack{}, fmt.Errorf("track %s doesn't exist", trackID)
	}
	if err := s.kv.Put(ctx, storeKey, *track); err != nil {
		return spotify.FullTrack{}, fmt.Errorf("storing track %s: %w", trackID, err)
	}
	return *track, nil
}

func spotifyOpts(opts paginator.PageOpts) *spotify.Options {
	return &spotify.Options{Limit: &opts.Limit, Offset: &opts.Offset}
}
