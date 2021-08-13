package recommendations

import (
	"fmt"

	"github.com/kristofferostlund/recommendli/pkg/logging"
	"github.com/kristofferostlund/recommendli/pkg/spotifypaginator"
	"github.com/zmb3/spotify"
)

type SpotifyProvider interface{}

type CacheProvider interface {
	Get(key string, out interface{}) error
	Put(key string, out interface{}) error
}

type ServiceFactory struct {
	log logging.Logger
}

func NewServiceFactory(log logging.Logger) *ServiceFactory {
	return &ServiceFactory{log: log}
}

func (s *ServiceFactory) NewService(spotifyClient spotify.Client) *Service {
	return &Service{log: s.log, spotify: spotifyClient}
}

type Service struct {
	log     logging.Logger
	spotify spotify.Client
}

func (s *Service) listPlaylists(usr spotify.User) ([]spotify.SimplePlaylist, error) {
	playlists := make([]spotify.SimplePlaylist, 0)
	paginator := spotifypaginator.New(
		spotifypaginator.PageSize(50),
		spotifypaginator.ProgressReporter(func(currentCount, totalCount, currentPage int) {
			s.log.Info("listing simple playlists", "user", usr.DisplayName, "count", currentCount, "total", totalCount, "page", currentPage)
		}),
	)
	if err := paginator.Run(func(opts *spotify.Options, next spotifypaginator.NextFunc) (*spotifypaginator.NextResult, error) {
		r, err := s.spotify.GetPlaylistsForUserOpt(usr.ID, opts)
		if err != nil {
			return nil, fmt.Errorf("listing playlists for user %s: %w", usr.ID, err)
		}
		playlists = append(playlists, r.Playlists...)
		return next(len(playlists), r.Total), nil
	}); err != nil {
		return nil, err
	}

	return playlists, nil
}

func (s *Service) currentUser() (spotify.User, error) {
	usr, err := s.spotify.CurrentUser()
	if err != nil {
		return spotify.User{}, fmt.Errorf("getting current user: %w", err)
	}
	return usr.User, nil
}
