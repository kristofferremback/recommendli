package recommendations

import (
	"context"
	"regexp"
)

type DummyUserPreferenceProvider struct{}

func (d *DummyUserPreferenceProvider) GetLibraryPattern(ctx context.Context, userID string) (*regexp.Regexp, error) {
	return regexp.Compile(`^Metal \d+`)
}
