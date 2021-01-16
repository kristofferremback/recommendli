package spotify

import "github.com/zmb3/spotify"

func fromSimpleArtists(simpleArtists []spotify.SimpleArtist) []Artist {
	artists := []Artist{}
	for _, a := range simpleArtists {
		artists = append(artists, Artist{
			ID:   a.ID.String(),
			Name: a.Name,
		})
	}
	return artists
}
