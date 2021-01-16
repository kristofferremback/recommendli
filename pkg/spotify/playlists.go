package spotify

import (
	"fmt"

	"github.com/zmb3/spotify"
)

// TODO: Figure out how to resolve all tracks

type Playlist struct {
	ID         string
	Name       string
	SnapshotID string
	Tracks     []Track
	Size       int
}

func (p Playlist) Populated() bool {
	return p.Size == len(p.Tracks)
}

func (p Playlist) SpotifyID() spotify.ID {
	return spotify.ID(p.ID)
}

func fromSimplePlaylists(simplePlaylists []spotify.SimplePlaylist) []Playlist {
	playlists := []Playlist{}
	for _, s := range simplePlaylists {
		playlists = append(playlists, Playlist{
			ID:         string(s.ID),
			Name:       s.Name,
			SnapshotID: s.SnapshotID,
			Size:       int(s.Tracks.Total),
		})
	}
	return playlists
}

func (c *client) listPlaylists(user User) ([]Playlist, error) {
	playlists := []Playlist{}

	paginator := newPaginator(limit(50))
	err := paginator.Paginate(func(opts *spotify.Options) (int, int, error) {
		page, err := c.spotify.GetPlaylistsForUserOpt(user.ID, opts)
		if err != nil {
			return 0, 0, err
		}

		playlists = append(playlists, fromSimplePlaylists(page.Playlists)...)
		return page.Total, len(playlists), nil
	})
	if err != nil {
		return playlists, fmt.Errorf("Error listing playlist: %w", err)
	}

	return playlists, nil
}

func (c *client) populatePlaylist(playlist Playlist) (Playlist, error) {
	if playlist.Populated() {
		return playlist, nil
	}

	paginator := newPaginator(limit(100))
	err := paginator.Paginate(func(opts *spotify.Options) (int, int, error) {
		page, err := c.spotify.GetPlaylistTracksOpt(playlist.SpotifyID(), paginator.Options(), "")
		if err != nil {
			return 0, 0, err
		}

		playlist.Tracks = append(playlist.Tracks, fromPlaylistTracks(page.Tracks)...)
		return page.Total, len(playlist.Tracks), nil
	})
	if err != nil {
		return playlist, fmt.Errorf("Failed to populate playlist tracks: %w", err)
	}

	return playlist, nil
}
