package spotify

import (
	"errors"
	"net/http"
	"time"

	"github.com/zmb3/spotify"
)

var AuthenticationError error = errors.New("No authentication")

type Client interface {
	GetUser() (User, error)
	ListPlaylists(user User) ([]Playlist, error)
}

type Auth interface {
	Middleware() func(h http.Handler) http.Handler
	TokenCallbackHandler() http.HandlerFunc
	Redirect(w http.ResponseWriter, r *http.Request)
	GetClient(r *http.Request) (Client, error)
}

type User struct {
	ID   string
	Name string
}

func (u User) SpotifyID() spotify.ID {
	return spotify.ID(u.ID)
}

type Track struct {
	ID      string
	Name    string
	Artists []Artist
	Album   Album
}

func (t Track) SpotifyID() spotify.ID {
	return spotify.ID(t.ID)
}

type Artist struct {
	ID   string
	Name string
}

func (a Artist) SpotifyID() spotify.ID {
	return spotify.ID(a.ID)
}

type Album struct {
	ID          string
	Name        string
	Artists     []Artist
	ReleaseDate time.Time
	Kind        string
}

func (a Album) SpotifyID() spotify.ID {
	return spotify.ID(a.ID)
}
