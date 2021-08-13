package recommendations

import (
	"context"

	"github.com/kristofferostlund/recommendli/pkg/logging"
	"github.com/zmb3/spotify"
)

type SpotifyProvider interface {
	ListPlaylists(ctx context.Context, usr spotify.User) ([]spotify.SimplePlaylist, error)
	CurrentUser(ctx context.Context) (spotify.User, error)
}

type ServiceFactory struct {
	log logging.Logger
}

func NewServiceFactory(log logging.Logger) *ServiceFactory {
	return &ServiceFactory{log: log}
}

func (f *ServiceFactory) New(spotifyProvider SpotifyProvider) *Service {
	return &Service{log: f.log, spotify: spotifyProvider}
}

type Service struct {
	log     logging.Logger
	spotify SpotifyProvider
}

func (s *Service) ListPlaylistsForCurrentUser(ctx context.Context) ([]spotify.SimplePlaylist, error) {
	usr, err := s.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	playlists, err := s.spotify.ListPlaylists(ctx, usr)
	if err != nil {
		return nil, err
	}
	return playlists, nil
}

func (s *Service) GetCurrentUser(ctx context.Context) (spotify.User, error) {
	user, err := s.spotify.CurrentUser(ctx)
	if err != nil {
		return spotify.User{}, err
	}
	return user, nil
}
