package spotify

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"

	"github.com/kristofferostlund/recommendli/pkg/httphelpers"
	"github.com/kristofferostlund/recommendli/pkg/logging"
)

const (
	ContextToken = "SpotifyClient"

	CookieState        = "recommendli_authstate"
	CookieGoto         = "recommendli_goto"
	CookieSpotifyToken = "recommendli_spotifytoken"
)

type keyValueStore interface {
	Put(key string, value interface{}) error
	Get(key string, out interface{}) (bool, error)
}

type authService struct {
	authenticator spotify.Authenticator
	redirectURL   url.URL

	log logging.Logger
}

func NewAuthService(clientID, clientSecret string, redirectURL url.URL, log logging.Logger) *authService {
	authenticator := spotify.NewAuthenticator(
		redirectURL.String(),
		spotify.ScopeUserReadPrivate,
		spotify.ScopePlaylistReadPrivate,
		spotify.ScopePlaylistModifyPrivate,
		spotify.ScopePlaylistModifyPublic,
		spotify.ScopeUserTopRead,
		spotify.ScopeUserReadCurrentlyPlaying,
		spotify.ScopeUserReadPlaybackState,
	)
	authenticator.SetAuthInfo(clientID, clientSecret)

	return &authService{
		authenticator: authenticator,
		redirectURL:   redirectURL,
		log:           log,
	}
}

func (a *authService) TokenCallbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, _ := r.Cookie(CookieState)
		if c == nil || c.Value == "" {
			a.log.Error("Missing required cookie", fmt.Errorf("Missing required cookie %s", CookieState))
			httphelpers.InternalServerError(w)
			return
		}

		state, err := url.QueryUnescape(c.Value)
		httphelpers.ClearCookie(w, c)
		if err != nil {
			a.log.Error("Failed to escape state", err)
			httphelpers.InternalServerError(w)
			return
		}

		token, err := a.authenticator.Token(state, r)
		if err != nil {
			a.log.Error("Failed to get token", err)
			httphelpers.InternalServerError(w)
			return
		}

		tokenB, err := json.Marshal(token)
		if err != nil {
			a.log.Error("Failed to marshal token", err)
			httphelpers.InternalServerError(w)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     CookieSpotifyToken,
			Value:    base64.StdEncoding.EncodeToString(tokenB),
			Expires:  token.Expiry,
			Path:     "/",
			Secure:   true,
			HttpOnly: true,
			// @TODO: Read up on what cookie method to use so this is actually secure
			SameSite: http.SameSiteLaxMode,
		})

		gc, _ := r.Cookie(CookieGoto)
		if gc != nil && gc.Value != "" {
			redirectTo, err := url.QueryUnescape(gc.Value)
			httphelpers.ClearCookie(w, gc)
			if err != nil {
				a.log.Error("Failed to get token", err)
				httphelpers.InternalServerError(w)
				return
			}

			http.Redirect(w, r, redirectTo, http.StatusTemporaryRedirect)
			return
		}

		w.Write([]byte("OK"))
	}
}

func (a *authService) Middleware() func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, _ := r.Cookie(CookieSpotifyToken)
			if c != nil && c.Value != "" {
				token := &oauth2.Token{}
				decoded, _ := base64.StdEncoding.DecodeString(c.Value)
				err := json.Unmarshal(decoded, token)

				if err != nil {
					a.log.With("error", err).Warn("Failed to unmarshal token")
					a.Redirect(w, r)
					return
				}
				if !token.Valid() {
					a.Redirect(w, r)
					return
				}

				r = r.WithContext(context.WithValue(r.Context(), ContextToken, token))
				h.ServeHTTP(w, r)
				return
			}

			a.Redirect(w, r)
		})
	}
}

func (a *authService) GetClient(r *http.Request) (Client, error) {
	token, ok := r.Context().Value(ContextToken).(*oauth2.Token)
	if !ok {
		return nil, AuthenticationError
	}
	spotifyClient := a.authenticator.NewClient(token)
	return &client{
		spotify: spotifyClient,
		log:     a.log,
	}, nil
}

func (a *authService) Redirect(w http.ResponseWriter, r *http.Request) {
	state := uuid.NewV4().String()

	http.SetCookie(w, &http.Cookie{
		Name:     CookieState,
		Value:    url.QueryEscape(state),
		Expires:  time.Now().Add(time.Hour),
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		// @TODO: Read up on what cookie method to use so this is actually secure
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     CookieGoto,
		Value:    url.QueryEscape(r.URL.String()),
		Expires:  time.Now().Add(time.Hour),
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		// @TODO: Read up on what cookie method to use so this is actually secure
		SameSite: http.SameSiteLaxMode,
	})

	authURL := a.authenticator.AuthURL(state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}
