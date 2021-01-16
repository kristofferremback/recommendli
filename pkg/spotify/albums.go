package spotify

import "github.com/zmb3/spotify"

func fromSimpleAlbum(simpleAlbum spotify.SimpleAlbum) Album {
	return Album{
		ID:          simpleAlbum.ID.String(),
		Name:        simpleAlbum.Name,
		Artists:     fromSimpleArtists(simpleAlbum.Artists),
		ReleaseDate: simpleAlbum.ReleaseDateTime(),
		Kind:        simpleAlbum.AlbumType,
	}
}
