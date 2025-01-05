package spotifyutil

import "github.com/zmb3/spotify"

func SimplePlaylistHasChanged(a, b spotify.SimplePlaylist) bool {
	return a.SnapshotID != b.SnapshotID || int(a.Tracks.Total) != int(b.Tracks.Total)
}
