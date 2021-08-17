package recommendations

import (
	"context"
	"fmt"
	"regexp"

	"github.com/kristofferostlund/recommendli/pkg/logging"
	"github.com/zmb3/spotify"
)

type KeyValueStore interface {
	Get(ctx context.Context, key string, out interface{}) (bool, error)
	Put(ctx context.Context, key string, data interface{}) error
}

type SpotifyProvider interface {
	ListPlaylists(ctx context.Context, userID string) ([]spotify.SimplePlaylist, error)
	GetPlaylist(ctx context.Context, playlistID string) (spotify.FullPlaylist, error)
	PopulatePlaylists(ctx context.Context, simplePlaylists []spotify.SimplePlaylist) ([]spotify.FullPlaylist, error)
	CurrentUser(ctx context.Context) (spotify.User, error)
	CurrentTrack(ctx context.Context) (spotify.FullTrack, bool, error)
}

type UserPreferenceProvider interface {
	GetLibraryPattern(ctx context.Context, userID string) (*regexp.Regexp, error)
}

type ServiceFactory struct {
	log             logging.Logger
	store           KeyValueStore
	userPreferences UserPreferenceProvider
}

func NewServiceFactory(log logging.Logger, store KeyValueStore, userPreferences UserPreferenceProvider) *ServiceFactory {
	return &ServiceFactory{log: log, store: store, userPreferences: userPreferences}
}

func (f *ServiceFactory) New(spotifyProvider SpotifyProvider) *Service {
	return &Service{log: f.log, store: f.store, userPreferences: f.userPreferences, spotify: spotifyProvider}
}

type Service struct {
	log             logging.Logger
	store           KeyValueStore
	userPreferences UserPreferenceProvider
	spotify         SpotifyProvider
}

func (s *Service) ListPlaylistsForCurrentUser(ctx context.Context) ([]spotify.SimplePlaylist, error) {
	usr, err := s.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	return s.spotify.ListPlaylists(ctx, usr.ID)
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

	matching := filterMatchingNames(re, playlists)
	if len(matching) == 0 {
		return nil, nil
	}

	return s.spotify.PopulatePlaylists(ctx, matching)
}

func (s *Service) GetPlaylist(ctx context.Context, playlistID string) (spotify.FullPlaylist, error) {
	return s.spotify.GetPlaylist(ctx, playlistID)
}

func (s *Service) GetCurrentUser(ctx context.Context) (spotify.User, error) {
	return s.spotify.CurrentUser(ctx)
}

type ErrNoCurrentTrack struct {
	usr spotify.User
}

func (err ErrNoCurrentTrack) Error() string {
	return fmt.Sprintf("user %s must listen to music", err.usr.DisplayName)
}

func (s *Service) CheckPlayingTrackInLibrary(ctx context.Context) (spotify.FullTrack, []spotify.SimplePlaylist, error) {
	usr, err := s.GetCurrentUser(ctx)
	if err != nil {
		return spotify.FullTrack{}, nil, err
	}

	currentTrack, isPlaying, err := s.spotify.CurrentTrack(ctx)
	if err != nil {
		return spotify.FullTrack{}, nil, fmt.Errorf("checking track current user is playing: %w", err)
	}
	if !isPlaying {
		return spotify.FullTrack{}, nil, ErrNoCurrentTrack{usr: usr}
	}

	playlists, err := s.spotify.ListPlaylists(ctx, usr.ID)
	if err != nil {
		return spotify.FullTrack{}, nil, fmt.Errorf("listing user playlists checking if track is in library: %w", err)
	}
	re, err := s.userPreferences.GetLibraryPattern(ctx, usr.ID)
	if err != nil {
		return spotify.FullTrack{}, nil, fmt.Errorf("checking if track is in library: %w", err)
	}
	libraryPlaylists := filterMatchingNames(re, playlists)

	indexedLibrary, err := s.getStoredTrackPlaylistIndex(ctx, usr, libraryPlaylists)
	if err != nil {
		return spotify.FullTrack{}, nil, fmt.Errorf("populating base playlists when checking if track is in library: %w", err)
	}

	s.log.Info("tracks fully listed", "unique song count", len(indexedLibrary.Tracks), "playlist count", len(indexedLibrary.Playlists))

	if pls, found := indexedLibrary.Lookup(currentTrack); found {
		playlistNames := make([]string, 0, len(pls))
		for _, p := range pls {
			playlistNames = append(playlistNames, p.Name)
		}
		s.log.Info("current track already in library", "track", indexedLibrary.Key(currentTrack), "playlists", playlistNames)
		return currentTrack, pls, nil
	}

	s.log.Info("current track is new", "track", indexedLibrary.Key(currentTrack))
	return currentTrack, nil, nil
}
