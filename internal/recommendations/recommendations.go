package recommendations

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/kristofferostlund/recommendli/pkg/keyvaluestore"
	"github.com/zmb3/spotify"
)

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
	store           keyvaluestore.KV
	userPreferences UserPreferenceProvider
	trackIndex      TrackIndex
}

func NewServiceFactory(store keyvaluestore.KV, userPreferences UserPreferenceProvider, trackIndex TrackIndex) *ServiceFactory {
	return &ServiceFactory{
		store:           store,
		userPreferences: userPreferences,
		trackIndex:      trackIndex,
	}
}

func (f *ServiceFactory) New(spotifyProvider SpotifyProvider) *service {
	return &service{
		store:           f.store,
		userPreferences: f.userPreferences,
		spotify:         spotifyProvider,
		trackIndex:      f.trackIndex,
	}
}

type service struct {
	store           keyvaluestore.KV
	userPreferences UserPreferenceProvider
	spotify         SpotifyProvider
	trackIndex      TrackIndex
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

func (s *service) ListPlaylistsForCurrentUser(ctx context.Context) ([]spotify.SimplePlaylist, error) {
	usr, err := s.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	return s.spotify.ListPlaylists(ctx, usr.ID)
}

func (s *service) GetCurrentUsersPlaylistMatchingPattern(ctx context.Context, pattern string) ([]spotify.FullPlaylist, error) {
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

func (s *service) GetPlaylist(ctx context.Context, playlistID string) (spotify.FullPlaylist, error) {
	return s.spotify.GetPlaylist(ctx, playlistID)
}

func (s *service) GetCurrentUser(ctx context.Context) (spotify.User, error) {
	return s.spotify.CurrentUser(ctx)
}

type ErrNoCurrentTrack struct {
	usr spotify.User
}

func (err ErrNoCurrentTrack) Error() string {
	return fmt.Sprintf("user %s must listen to music", err.usr.DisplayName)
}

func (s *service) GetCurrentlyPlayingTrackAlbum(ctx context.Context) (spotify.FullAlbum, error) {
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

func (s *service) GetCurrentTrack(ctx context.Context) (spotify.FullTrack, bool, error) {
	return s.spotify.CurrentTrack(ctx)
}

func (s *service) CheckPlayingTrackInLibrary(ctx context.Context) (spotify.FullTrack, []spotify.SimplePlaylist, error) {
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

	if err := s.prepareTrackIndexForUser(ctx, usr); err != nil {
		return spotify.FullTrack{}, nil, fmt.Errorf("getting track index: %w", err)
	}

	pls, err := s.trackIndex.Lookup(ctx, usr.ID, currentTrack.SimpleTrack)
	if err != nil {
		return spotify.FullTrack{}, nil, fmt.Errorf("looking up track in library: %w", err)
	}
	if len(pls) > 0 {
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

func (s *service) CreateDiscoveryPlaylist(ctx context.Context) (spotify.FullPlaylist, error) {
	return s.generateDiscoveryPlaylist(ctx, false)
}

func (s *service) DryRunDiscoveryPlaylist(ctx context.Context) (spotify.FullPlaylist, error) {
	return s.generateDiscoveryPlaylist(ctx, true)
}

func (s *service) GetIndexSummary(ctx context.Context) (IndexSummary, error) {
	usr, err := s.GetCurrentUser(ctx)
	if err != nil {
		return IndexSummary{}, fmt.Errorf("getting user: %w", err)
	}

	if err := s.prepareTrackIndexForUser(ctx, usr); err != nil {
		return IndexSummary{}, fmt.Errorf("getting track index for user: %w", err)
	}

	summary, err := s.trackIndex.Summarize(ctx, usr.ID)
	if err != nil {
		return IndexSummary{}, fmt.Errorf("getting index summary: %w", err)
	}

	return summary, nil
}

func (s *service) prepareTrackIndexForUser(ctx context.Context, usr spotify.User) error {
	playlists, err := s.spotify.ListPlaylists(ctx, usr.ID)
	if err != nil {
		return fmt.Errorf("listing user playlists: %w", err)
	}
	prefs, err := s.userPreferences.GetPreferences(ctx, usr.ID)
	if err != nil {
		return fmt.Errorf("getting user prefences: %w", err)
	}
	libraryPlaylists := filterSimplePlaylist(playlists, func(p spotify.SimplePlaylist) bool {
		return prefs.IsLibraryPlaylistName(p.Name)
	})

	if err := s.ensureTrackIndexSynced(ctx, usr.ID, libraryPlaylists); err != nil {
		return fmt.Errorf("checking if track index needs sync: %w", err)
	}

	return nil
}

func (s *service) generateDiscoveryPlaylist(ctx context.Context, dryRun bool) (spotify.FullPlaylist, error) {
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

	if err := s.ensureTrackIndexSynced(ctx, usr.ID, libraryPlaylists); err != nil {
		return spotify.FullPlaylist{}, fmt.Errorf("checking if track index needs sync: %w", err)
	}

	populatedDiscovery, err := s.spotify.PopulatePlaylists(ctx, discoveryPlaylists)
	if err != nil {
		return spotify.FullPlaylist{}, fmt.Errorf("populating discovery playlists when generating discovery playlist: %w", err)
	}

	slog.DebugContext(ctx, "discovery playlists fully listed", "unique song count", len(uniqueTracks(tracksFor(populatedDiscovery))), "playlist count", len(populatedDiscovery))
	candidates := make([]spotify.FullTrack, 0)
	for _, t := range uniqueTracks(tracksFor(populatedDiscovery)) {
		has, err := s.trackIndex.Has(ctx, usr.ID, t.SimpleTrack)
		if err != nil {
			return spotify.FullPlaylist{}, fmt.Errorf("checking if track is in library when generating discovery playlist: %w", err)
		}

		slog.DebugContext(ctx, "candidate track", "track", stringifyTrack(t.SimpleTrack), "in_library", has)
		if !has {
			candidates = append(candidates, t)
		}
	}

	scores, err := s.scoreTracks(ctx, usr.ID, candidates)
	if err != nil {
		return spotify.FullPlaylist{}, fmt.Errorf("getting most relevant versions of tracks when generating discovery playlist: %w", err)
	}

	sort.SliceStable(scores, func(i, j int) bool {
		return scores[i].calculate(prefs) > scores[j].calculate(prefs)
	})
	tracks := make([]spotify.FullTrack, 0)
	for _, s := range scores {
		keep := s.keep(prefs)
		slog.DebugContext(ctx, "track score", "track", stringifyTrack(s.track.SimpleTrack), "score", s.calculate(prefs), "keep", keep)
		if keep {
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
