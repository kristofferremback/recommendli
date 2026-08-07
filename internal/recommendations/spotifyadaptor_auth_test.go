package recommendations

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/zmb3/spotify"
)

func TestSpotifyAuthRequestsPlaybackScopes(t *testing.T) {
	auth := newTestAuthAdaptor(t)
	authURL, err := url.Parse(auth.authenticator.AuthURL("state"))
	if err != nil {
		t.Fatalf("parsing auth URL: %v", err)
	}
	scopes := strings.Fields(authURL.Query().Get("scope"))

	for _, required := range []string{
		spotify.ScopeUserReadPlaybackState,
		spotify.ScopeUserModifyPlaybackState,
		spotify.ScopeUserReadRecentlyPlayed,
	} {
		if !containsString(scopes, required) {
			t.Fatalf("requested scopes %v do not contain %q", scopes, required)
		}
	}
}

func TestSpotifyAuthVersionForcesOldSessionsToReauthorize(t *testing.T) {
	auth := newTestAuthAdaptor(t)
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/recommendations/v1/playback", nil)
	req.AddCookie(&http.Cookie{Name: CookieSpotifyToken, Value: "old-token"})
	res := httptest.NewRecorder()

	auth.Middleware()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("request with old auth version reached protected handler")
	})).ServeHTTP(res, req)

	if res.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusTemporaryRedirect)
	}
	if location := res.Header().Get("Location"); !strings.Contains(location, "accounts.spotify.com/authorize") {
		t.Fatalf("Location = %q, want Spotify authorization URL", location)
	}
}

func newTestAuthAdaptor(t *testing.T) *AuthAdaptor {
	t.Helper()
	redirect, err := url.Parse("http://127.0.0.1/recommendations/v1/spotify/auth/callback")
	if err != nil {
		t.Fatal(err)
	}
	uiRedirect, err := url.Parse("http://127.0.0.1/recommendations/v1/spotify/auth/ui-redirect")
	if err != nil {
		t.Fatal(err)
	}
	return NewSpotifyAuthAdaptor("client-id", "client-secret", *redirect, *uiRedirect)
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
