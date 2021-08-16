package recommendations

import (
	"context"
	"fmt"

	"github.com/kristofferostlund/recommendli/pkg/ctxhelper"
	"github.com/kristofferostlund/recommendli/pkg/logging"
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
	p, err := s.spotify.PlayerCurrentlyPlaying()
	if err != nil {
		return spotify.FullTrack{}, false, fmt.Errorf("getting currently playing track: %w", err)
	}
	if !p.Playing {
		return spotify.FullTrack{}, false, nil
	}
	return *p.Item, true, nil
}
