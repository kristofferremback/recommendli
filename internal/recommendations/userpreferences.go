package recommendations

import (
	"context"
	"fmt"
	"regexp"
)

type UserPreferenceProviderImpl struct {
	kv KeyValueStore
}

type userPreferencesDTO struct {
	LibraryPattern         string         `json:"library_pattern"`
	DiscoveryPlaylistNames []string       `json:"discovery_playlist_names"`
	WeightedWords          map[string]int `json:"weighted_words"`
	MinimumAlbumSize       int            `json:"minimum_album_size"`
}

func (u userPreferencesDTO) UserPreferences() (UserPreferences, error) {
	libPattern, err := regexp.Compile(u.LibraryPattern)
	if err != nil {
		return UserPreferences{}, fmt.Errorf("compiling library pattern: %w", err)
	}
	return UserPreferences{
		LibraryPattern:         libPattern,
		DiscoveryPlaylistNames: u.DiscoveryPlaylistNames,
		WeightedWords:          u.WeightedWords,
		MinimumAlbumSize:       u.MinimumAlbumSize,
	}, nil
}

func userPreferencesDTOFor(prefs UserPreferences) userPreferencesDTO {
	return userPreferencesDTO{
		LibraryPattern:         prefs.LibraryPattern.String(),
		DiscoveryPlaylistNames: prefs.DiscoveryPlaylistNames,
		WeightedWords:          prefs.WeightedWords,
		MinimumAlbumSize:       prefs.MinimumAlbumSize,
	}
}

func NewUserPreferenceProvider(kv KeyValueStore) *UserPreferenceProviderImpl {
	return &UserPreferenceProviderImpl{kv: kv}
}

func (u *UserPreferenceProviderImpl) Get(ctx context.Context, userID string) (UserPreferences, error) {
	var prefs UserPreferences
	found, err := u.kv.Get(ctx, u.key(userID), &prefs)
	if err != nil {
		return UserPreferences{}, fmt.Errorf("getting user preferencess: %w", err)
	}
	if !found {
		return u.defaultUserPreferences(), nil
	}

	return prefs, nil
}

func (u *UserPreferenceProviderImpl) Set(ctx context.Context, userID string, prefs UserPreferences) error {
	if err := u.kv.Put(ctx, u.key(userID), prefs); err != nil {
		return fmt.Errorf("setting user preferences: %w", err)
	}
	return nil
}

func (UserPreferenceProviderImpl) key(userID string) string {
	return fmt.Sprintf("user-preferences-%s", userID)
}

func (UserPreferenceProviderImpl) defaultUserPreferences() UserPreferences {
	// return UserPreferences{
	// 	LibraryPattern:                   regexp.MustCompile(".*"),
	// 	DiscoveryPlaylistNames:           []string{"Release Radar", "Discover Weekly"},
	// 	WeightedWords:                    make(map[string]int),
	// 	MinimumAlbumSize:                 0,
	// }

	// legacy default.
	return UserPreferences{
		LibraryPattern:         regexp.MustCompile(`^Metal \d+`),
		DiscoveryPlaylistNames: []string{"Release Radar", "Discover Weekly"},
		WeightedWords: map[string]int{
			"instrumental": -50,
			"acoustic":     -30,
			"re-imagined":  -30,
			"remix":        -30,
		},
		MinimumAlbumSize: 4,
	}
}
