package recommendations

import (
	"context"
	"fmt"
	"regexp"

	"github.com/zmb3/spotify"
)

func (s *Service) getStoredTrackPlaylistIndex(ctx context.Context, usr spotify.User, simplePlaylists []spotify.SimplePlaylist) (*TrackPlaylistIndex, error) {
	storeKey := fmt.Sprintf("cache_track-playlist-index_%s", usr.ID)
	s.log.Debug("checking stored track playlist index", "user", usr.DisplayName, "key", storeKey)
	var index *TrackPlaylistIndex
	found, err := s.store.Get(ctx, storeKey, &index)
	if err != nil {
		return nil, err
	}
	if found && index.MatchesSimpleLaylists(simplePlaylists) {
		s.log.Debug("stored track playlist index", "user", usr.DisplayName, "key", storeKey)
		return index, nil
	}

	s.log.Debug("updating stored track playlist index", "user", usr.DisplayName, "key", storeKey)
	populatedLibrary, err := s.spotify.PopulatePlaylists(ctx, simplePlaylists)
	if err != nil {
		return nil, err
	}
	index = NewTrackPlaylistIndexFromFullPlaylists(populatedLibrary)
	if err := s.store.Put(ctx, storeKey, &index); err != nil {
		return nil, err
	}
	s.log.Debug("stored track playlist index updated", "user", usr.DisplayName, "key", storeKey)

	return index, nil
}

func filterMatchingNames(re *regexp.Regexp, playlists []spotify.SimplePlaylist) []spotify.SimplePlaylist {
	libraryPlaylists := make([]spotify.SimplePlaylist, 0)
	for _, p := range playlists {
		if re.MatchString(p.Name) {
			libraryPlaylists = append(libraryPlaylists, p)
		}
	}
	return libraryPlaylists
}
