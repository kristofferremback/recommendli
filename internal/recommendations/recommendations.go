package recommendations

import (
	"context"
	"fmt"
	"regexp"

	"github.com/kristofferostlund/recommendli/pkg/logging"
	"github.com/zmb3/spotify"
)

type SpotifyProvider interface {
	ListPlaylists(ctx context.Context, usr spotify.User) ([]spotify.SimplePlaylist, error)
	GetPlaylist(ctx context.Context, playlistID string) (spotify.FullPlaylist, error)
	PopulatePlaylists(ctx context.Context, simplePlaylists []spotify.SimplePlaylist) ([]spotify.FullPlaylist, error)
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
	return s.spotify.ListPlaylists(ctx, usr)
}

func (s *Service) GetCurrentUsersPlaylistMatchingPattern(ctx context.Context, pattern string) ([]spotify.FullPlaylist, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("getting current user's playlists: %w", err)
	}
	s.log.Debug("finding playlists matching pattern", "pattern", pattern)
	playlists, err := s.ListPlaylistsForCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	matching := make([]spotify.SimplePlaylist, 0)
	for _, p := range playlists {
		if re.MatchString(p.Name) {
			matching = append(matching, p)
		}
	}
	return s.spotify.PopulatePlaylists(ctx, matching)
}

func (s *Service) GetPlaylist(ctx context.Context, playlistID string) (spotify.FullPlaylist, error) {
	return s.spotify.GetPlaylist(ctx, playlistID)
}

func (s *Service) GetCurrentUser(ctx context.Context) (spotify.User, error) {
	return s.spotify.CurrentUser(ctx)
}
