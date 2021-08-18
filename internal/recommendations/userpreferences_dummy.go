package recommendations

import (
	"context"
	"regexp"
)

type DummyUserPreferenceProvider struct {
	prefs UserPreferences
}

func NewDummyUserPreferenceProvider() *DummyUserPreferenceProvider {
	return &DummyUserPreferenceProvider{
		prefs: UserPreferences{
			LibraryPattern:         regexp.MustCompile(`^Metal \d+`),
			DiscoveryPlaylistNames: []string{"Release Radar", "Discover Weekly"},
			WeightedWords: map[string]int{
				"instrumental": -50,
				"acoustic":     -30,
				"re-imagined":  -30,
				"remix":        -30,
			},
			MinimumAlbumSize:                 3,
			RecommendationPlaylistNamePrefix: "recommendli",
		},
	}
}

func (d *DummyUserPreferenceProvider) GetPreferences(ctx context.Context, userID string) (UserPreferences, error) {
	return d.prefs, nil
}
