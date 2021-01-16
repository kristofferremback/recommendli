package spotify

import "github.com/zmb3/spotify"

func fromPlaylistTracks(plTracks []spotify.PlaylistTrack) []Track {
	tracks := []Track{}
	for _, t := range plTracks {
		tracks = append(tracks, Track{
			ID:      t.Track.ID.String(),
			Name:    t.Track.Name,
			Album:   fromSimpleAlbum(t.Track.Album),
			Artists: fromSimpleArtists(t.Track.Artists),
		})
	}

	return tracks
}
