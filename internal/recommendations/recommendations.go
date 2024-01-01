package recommendations

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/zmb3/spotify"

	"github.com/kristofferostlund/recommendli/pkg/maputil"
	"github.com/kristofferostlund/recommendli/pkg/sortby"
)

type KeyValueStore interface {
	Get(ctx context.Context, key string, out interface{}) (bool, error)
	Put(ctx context.Context, key string, data interface{}) error
}

type SpotifyProvider interface {
	ListPlaylists(ctx context.Context, userID string) ([]spotify.SimplePlaylist, error)
	GetPlaylist(ctx context.Context, playlistID string) (spotify.FullPlaylist, error)
	PopulatePlaylists(ctx context.Context, simplePlaylists []spotify.SimplePlaylist) ([]spotify.FullPlaylist, error)
	CreatePlaylist(ctx context.Context, userID, name string, trackIDs []string) (spotify.FullPlaylist, error)
	SetPlaylistTracks(ctx context.Context, playlistID string, trackIDs []string) (spotify.FullPlaylist, error)
	TruncatePlaylist(ctx context.Context, playlistID, snapshotID string) error
	CurrentUser(ctx context.Context) (spotify.User, error)
	CurrentTrack(ctx context.Context) (spotify.FullTrack, bool, error)
	GetAlbum(ctx context.Context, albumID string) (spotify.FullAlbum, error)
	GetAlbums(ctx context.Context, albumIDs []string) ([]spotify.FullAlbum, error)
	ListArtistAlbums(ctx context.Context, artistID string) ([]spotify.SimpleAlbum, error)
	GetTrack(ctx context.Context, trackID string) (spotify.FullTrack, error)
}

type UserPreferenceProvider interface {
	GetPreferences(ctx context.Context, userID string) (UserPreferences, error)
}

type UserPreferences struct {
	LibraryPattern                   *regexp.Regexp
	DiscoveryPlaylistNames           []string
	WeightedWords                    map[string]int
	MinimumAlbumSize                 int
	RecommendationPlaylistNamePrefix string
}

func (u UserPreferences) IsDiscoveryPlaylistName(name string) bool {
	return stringsContain(u.DiscoveryPlaylistNames, name)
}

func (u UserPreferences) IsLibraryPlaylistName(name string) bool {
	return !u.IsDiscoveryPlaylistName(name) && u.LibraryPattern.MatchString(name)
}

func (u UserPreferences) RecommendationPlaylistName(kind string, now time.Time) string {
	return fmt.Sprintf("%s %s %s", u.RecommendationPlaylistNamePrefix, kind, now.Format("2006-01-02"))
}

type ServiceFactory struct {
	store           KeyValueStore
	userPreferences UserPreferenceProvider
}

func NewServiceFactory(store KeyValueStore, userPreferences UserPreferenceProvider) *ServiceFactory {
	return &ServiceFactory{store: store, userPreferences: userPreferences}
}

func (f *ServiceFactory) New(spotifyProvider SpotifyProvider) *Service {
	return &Service{store: f.store, userPreferences: f.userPreferences, spotify: spotifyProvider}
}

type Service struct {
	store           KeyValueStore
	userPreferences UserPreferenceProvider
	spotify         SpotifyProvider
}

type score struct {
	track          spotify.FullTrack
	album          spotify.FullAlbum
	artistRelevace int
}

func (s score) keep(prefs UserPreferences) bool {
	return len(s.album.Tracks.Tracks) >= prefs.MinimumAlbumSize
}

func (s score) calculate(prefs UserPreferences) int {
	value := 0
	for word, penalty := range prefs.WeightedWords {
		if strings.Contains(strings.ToLower(s.track.Name), strings.ToLower(word)) {
			value += penalty
		}
	}
	return value + s.artistRelevace + s.album.ReleaseDateTime().Year() - 2000 + s.album.Tracks.Total
}

func (s *Service) ListPlaylistsForCurrentUser(ctx context.Context) ([]spotify.SimplePlaylist, error) {
	usr, err := s.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	return s.spotify.ListPlaylists(ctx, usr.ID)
}

func (s *Service) GetCurrentUsersPlaylistMatchingPattern(ctx context.Context, pattern string) ([]spotify.FullPlaylist, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("getting current user's playlists: %w", err)
	}
	slog.Debug("finding playlists matching pattern", "pattern", pattern)
	playlists, err := s.ListPlaylistsForCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	matching := filterSimplePlaylist(playlists, func(p spotify.SimplePlaylist) bool {
		return re.MatchString(p.Name)
	})
	if len(matching) == 0 {
		return nil, nil
	}

	return s.spotify.PopulatePlaylists(ctx, matching)
}

func (s *Service) GetPlaylist(ctx context.Context, playlistID string) (spotify.FullPlaylist, error) {
	return s.spotify.GetPlaylist(ctx, playlistID)
}

func (s *Service) GetCurrentUser(ctx context.Context) (spotify.User, error) {
	return s.spotify.CurrentUser(ctx)
}

type ErrNoCurrentTrack struct {
	usr spotify.User
}

func (err ErrNoCurrentTrack) Error() string {
	return fmt.Sprintf("user %s must listen to music", err.usr.DisplayName)
}

func (s *Service) GetCurrentlyPlayingTrackAlbum(ctx context.Context) (spotify.FullAlbum, error) {
	usr, err := s.GetCurrentUser(ctx)
	if err != nil {
		return spotify.FullAlbum{}, err
	}

	currentTrack, isPlaying, err := s.spotify.CurrentTrack(ctx)
	if err != nil {
		return spotify.FullAlbum{}, fmt.Errorf("checking track current user is playing: %w", err)
	} else if !isPlaying {
		return spotify.FullAlbum{}, ErrNoCurrentTrack{usr: usr}
	}

	return s.albumForTrack(ctx, currentTrack)
}

func (s *Service) GetCurrentTrack(ctx context.Context) (spotify.FullTrack, bool, error) {
	return s.spotify.CurrentTrack(ctx)
}

func (s *Service) CheckPlayingTrackInLibrary(ctx context.Context) (spotify.FullTrack, []spotify.SimplePlaylist, error) {
	usr, err := s.GetCurrentUser(ctx)
	if err != nil {
		return spotify.FullTrack{}, nil, err
	}

	currentTrack, isPlaying, err := s.spotify.CurrentTrack(ctx)
	if err != nil {
		return spotify.FullTrack{}, nil, fmt.Errorf("checking track current user is playing: %w", err)
	} else if !isPlaying {
		return spotify.FullTrack{}, nil, ErrNoCurrentTrack{usr: usr}
	}

	indexedLibrary, err := s.trackIndexFor(ctx, usr)
	if err != nil {
		return spotify.FullTrack{}, nil, fmt.Errorf("getting track index: %w", err)
	}

	slog.InfoContext(ctx, "tracks fully listed", "unique song count", len(indexedLibrary.Tracks), "playlist count", len(indexedLibrary.Playlists))

	if pls, found := indexedLibrary.Lookup(currentTrack); found {
		playlistNames := make([]string, 0, len(pls))
		for _, p := range pls {
			playlistNames = append(playlistNames, p.Name)
		}
		slog.InfoContext(ctx, "current track already in library", "track", stringifyTrack(currentTrack.SimpleTrack), "playlists", playlistNames)
		return currentTrack, pls, nil
	}

	slog.InfoContext(ctx, "current track is new", "track", stringifyTrack(currentTrack.SimpleTrack))
	return currentTrack, nil, nil
}

func (s *Service) CreateDiscoveryPlaylist(ctx context.Context) (spotify.FullPlaylist, error) {
	return s.generateDiscoveryPlaylist(ctx, false)
}

func (s *Service) DryRunDiscoveryPlaylist(ctx context.Context) (spotify.FullPlaylist, error) {
	return s.generateDiscoveryPlaylist(ctx, true)
}

func (s *Service) GetIndexSummary(ctx context.Context) (IndexSummary, error) {
	usr, err := s.GetCurrentUser(ctx)
	if err != nil {
		return IndexSummary{}, fmt.Errorf("getting user: %w", err)
	}

	indexedLibrary, err := s.trackIndexFor(ctx, usr)
	if err != nil {
		return IndexSummary{}, fmt.Errorf("getting track index for user: %w", err)
	}

	playlists := maputil.Values(indexedLibrary.Playlists)
	sort.Slice(playlists, func(i, j int) bool {
		return sortby.PaddedNumbers(playlists[i].Name, playlists[j].Name, 10, true)
	})

	summary := IndexSummary{
		UniqueTrackCount: len(indexedLibrary.Tracks),
		PlaylistCount:    len(indexedLibrary.Playlists),
		Playlists:        playlists,
	}

	return summary, nil
}

func (s *Service) trackIndexFor(ctx context.Context, usr spotify.User) (*TrackPlaylistIndex, error) {
	playlists, err := s.spotify.ListPlaylists(ctx, usr.ID)
	if err != nil {
		return nil, fmt.Errorf("listing user playlists: %w", err)
	}
	prefs, err := s.userPreferences.GetPreferences(ctx, usr.ID)
	if err != nil {
		return nil, fmt.Errorf("getting user prefences: %w", err)
	}
	libraryPlaylists := filterSimplePlaylist(playlists, func(p spotify.SimplePlaylist) bool {
		return prefs.IsLibraryPlaylistName(p.Name)
	})

	indexedLibrary, err := s.getStoredTrackPlaylistIndex(ctx, usr, libraryPlaylists)
	if err != nil {
		return nil, fmt.Errorf("populating library playlists when checking if track is in library: %w", err)
	}
	return indexedLibrary, nil
}

func (s *Service) generateDiscoveryPlaylist(ctx context.Context, dryRun bool) (spotify.FullPlaylist, error) {
	usr, err := s.GetCurrentUser(ctx)
	if err != nil {
		return spotify.FullPlaylist{}, err
	}

	playlists, err := s.spotify.ListPlaylists(ctx, usr.ID)
	if err != nil {
		return spotify.FullPlaylist{}, fmt.Errorf("listing user playlists generating discovery playlist: %w", err)
	}
	prefs, err := s.userPreferences.GetPreferences(ctx, usr.ID)
	if err != nil {
		return spotify.FullPlaylist{}, fmt.Errorf("getting user prefences: %w", err)
	}

	discoveryPlaylists := filterSimplePlaylist(playlists, func(p spotify.SimplePlaylist) bool {
		return prefs.IsDiscoveryPlaylistName(p.Name)
	})
	libraryPlaylists := filterSimplePlaylist(playlists, func(p spotify.SimplePlaylist) bool {
		return prefs.IsLibraryPlaylistName(p.Name)
	})

	indexedLibrary, err := s.getStoredTrackPlaylistIndex(ctx, usr, libraryPlaylists)
	if err != nil {
		return spotify.FullPlaylist{}, fmt.Errorf("populating library playlists when generating discovery playlist: %w", err)
	}

	populatedDiscovery, err := s.spotify.PopulatePlaylists(ctx, discoveryPlaylists)
	if err != nil {
		return spotify.FullPlaylist{}, fmt.Errorf("populating discovery playlists when generating discovery playlist: %w", err)
	}
	candidates := filterTracks(uniqueTracks(tracksFor(populatedDiscovery)), func(t spotify.FullTrack) bool {
		return !indexedLibrary.Has(t)
	})
	scores, err := s.scoreTracks(ctx, candidates, countArtistTracks(simpleTrackMapToSlice(indexedLibrary.Tracks)))
	if err != nil {
		return spotify.FullPlaylist{}, fmt.Errorf("getting most relevant versions of tracks when generating discovery playlist: %w", err)
	}

	sort.SliceStable(scores, func(i, j int) bool {
		return scores[i].calculate(prefs) > scores[j].calculate(prefs)
	})
	tracks := make([]spotify.FullTrack, 0)
	for _, s := range scores {
		if s.keep(prefs) {
			tracks = append(tracks, s.track)
		}
	}

	playlistName := prefs.RecommendationPlaylistName("discovery", time.Now())
	if dryRun {
		dummy := dummyPlaylistFor(playlistName, tracks)
		slog.InfoContext(ctx, "recommendation complete, not creating playlist", "dryrun", dryRun, "playlist", dummy.Name, "tracks", printableTracks(tracksOf(dummy)), "track count", dummy.Tracks.Total)
		return dummy, nil
	}
	playlist, err := s.upsertPlaylistByName(ctx, playlists, usr.ID, playlistName, trackIDsOf(tracks))
	if err != nil {
		return spotify.FullPlaylist{}, fmt.Errorf("setting discovery playlist %s for user %s: %w", playlistName, usr.ID, err)
	}
	slog.InfoContext(ctx, "recommendation complete", "playlist", playlist.Name, "tracks", printableTracks(tracksOf(playlist)), "track count", playlist.Tracks.Total)
	return playlist, nil
}
