package spotify

import (
	"github.com/kristofferostlund/recommendli/pkg/logging"
	"github.com/zmb3/spotify"
)

// @TODO: Be smart
// var (
// 	artistFields   = "id,name"
// 	albumFields    = fmt.Sprintf("id,name,artists.items(%s)", artistFields)
// 	playlistFields = fmt.Sprintf("tracks.items(track(name,album(%s)))", albumFields)
// )

type client struct {
	spotify spotify.Client
	log     logging.Logger
}

func (c *client) GetUser() (User, error) {
	return c.getUser()
}

func (c *client) ListPlaylists(user User) ([]Playlist, error) {
	return c.listPlaylists(user)
}
