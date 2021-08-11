package recommendations

import (
	"github.com/zmb3/spotify"
)

type SpotifyProvider interface{}

type Service struct {
	spotify spotify.Client
}

type ServiceFactory struct{}

func NewServiceFactory() *ServiceFactory {
	return &ServiceFactory{}
}

func (s *ServiceFactory) NewService(spotifyClient spotify.Client) *Service {
	return &Service{spotify: spotifyClient}
}
