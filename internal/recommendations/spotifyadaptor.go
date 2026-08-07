package recommendations

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kristofferostlund/recommendli/pkg/ctxhelper"
	"github.com/kristofferostlund/recommendli/pkg/paginator"
	"github.com/zmb3/spotify"
)

type SpotifyAdaptor struct {
	spotify    spotify.Client
	http       *http.Client
	apiBaseURL string
	kv         KeyValueStore
}

type SpotifyAdaptorFactory struct {
	store KeyValueStore
}

func NewSpotifyProviderFactory(store KeyValueStore) *SpotifyAdaptorFactory {
	return &SpotifyAdaptorFactory{store: store}
}

func (f *SpotifyAdaptorFactory) New(spotifyClient spotify.Client, httpClients ...*http.Client) *SpotifyAdaptor {
	var httpClient *http.Client
	if len(httpClients) > 0 {
		httpClient = httpClients[0]
	}
	return &SpotifyAdaptor{
		spotify:    spotifyClient,
		http:       httpClient,
		apiBaseURL: spotifyAPIBaseURL,
		kv:         f.store,
	}
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
