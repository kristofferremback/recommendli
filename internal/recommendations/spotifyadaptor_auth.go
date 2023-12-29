package recommendations

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"

	"github.com/kristofferostlund/recommendli/pkg/slogutil"
	"github.com/kristofferostlund/recommendli/pkg/srv"
)

type ctxTokenType string

const (
	ctxTokenKey ctxTokenType = "SpotifyClient"

	CookieState        = "recommendli_authstate"
	CookieGoto         = "recommendli_goto"
	CookieSpotifyToken = "recommendli_spotifytoken"
)

var ErrNoAuthentication error = errors.New("no authentication found")

type AuthAdaptor struct {
	authenticator              spotify.Authenticator
	redirectURL, uiRedirectURL url.URL
}

func NewSpotifyAuthAdaptor(clientID, clientSecret string, redirectURL, uiRedirectURL url.URL) *AuthAdaptor {
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

	return &AuthAdaptor{
		authenticator: authenticator,
		redirectURL:   redirectURL,
		uiRedirectURL: uiRedirectURL,
	}
}

func (a *AuthAdaptor) Path() string {
	return a.redirectURL.Path
}

func (a *AuthAdaptor) UIRedirectPath() string {
	return a.uiRedirectURL.Path
}

func (a *AuthAdaptor) TokenCallbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		c, _ := r.Cookie(CookieState)
		if c == nil || c.Value == "" {
			slog.ErrorContext(ctx, "Missing required cookie", slogutil.Error(fmt.Errorf("missing required cookie %s", CookieState)))
			srv.InternalServerError(w)
			return
		}

		state, err := url.QueryUnescape(c.Value)
		srv.ClearCookie(w, c)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to escape state", slog.Any("error", err))
			srv.InternalServerError(w)
			return
		}

		token, err := a.authenticator.Token(state, r)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get token", slogutil.Error(err))
			srv.InternalServerError(w)
			return
		}

		tokenB, err := json.Marshal(token)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to marshal token", slogutil.Error(err))
			srv.InternalServerError(w)
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
			srv.ClearCookie(w, gc)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to get token", slogutil.Error(err))
				srv.InternalServerError(w)
				return
			}

			http.Redirect(w, r, redirectTo, http.StatusTemporaryRedirect)
			return
		}

		w.Write([]byte("OK"))
	}
}

func (a *AuthAdaptor) UIRedirectHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		redirectTo := r.URL.Query().Get("url")
		if redirectTo == "" {
			slog.WarnContext(ctx, "No url provided, cannot redirect client")
			srv.JSONError(w, errors.New("url is a required paramter"), srv.Status(400))
			return
		}
		a.redirect(w, r, redirectTo)
	})
}

func (a *AuthAdaptor) Middleware() func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			c, _ := r.Cookie(CookieSpotifyToken)
			if c != nil && c.Value != "" {
				token := &oauth2.Token{}
				decoded, _ := base64.StdEncoding.DecodeString(c.Value)
				err := json.Unmarshal(decoded, token)
				if err != nil {
					slog.WarnContext(ctx, "Failed to unmarshal token", slogutil.Error(err))
					a.redirect(w, r, r.URL.String())
					return
				}
				if !token.Valid() {
					a.redirect(w, r, r.URL.String())
					return
				}

				r = r.WithContext(context.WithValue(r.Context(), ctxTokenKey, token))
				h.ServeHTTP(w, r)
				return
			}

			a.redirect(w, r, r.URL.String())
		})
	}
}

func (a *AuthAdaptor) GetClient(r *http.Request) (spotify.Client, error) {
	token, ok := r.Context().Value(ctxTokenKey).(*oauth2.Token)
	if !ok {
		return spotify.Client{}, ErrNoAuthentication
	}
	client := a.authenticator.NewClient(token)
	client.AutoRetry = true
	return client, nil
}

func (a *AuthAdaptor) redirect(w http.ResponseWriter, r *http.Request, redirectBackTo string) {
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
		Value:    url.QueryEscape(redirectBackTo),
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
