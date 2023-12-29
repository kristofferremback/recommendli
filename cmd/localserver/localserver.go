package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/kristofferostlund/recommendli/internal/recommendations"
	"github.com/kristofferostlund/recommendli/pkg/keyvaluestore"
	"github.com/kristofferostlund/recommendli/pkg/logging"
	"github.com/kristofferostlund/recommendli/pkg/slogutil"
	"github.com/kristofferostlund/recommendli/pkg/srv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	defaultAddr = ":9999"
)

var defaultSpotifyRedirectHost = fmt.Sprintf("http://localhost%s", defaultAddr)

func main() {
	var (
		addr                = flag.String("addr", defaultAddr, "HTTP address")
		spotifyRedirectHost = flag.String("spotify-redirect-host", defaultSpotifyRedirectHost, "Spotify redirect host")
		logLevel            = flag.String("log-level", logging.LevelInfo.String(), "log level")

		clientID         = envString("SPOTIFY_ID", "")
		clientSecret     = envString("SPOTIFY_SECRET", "")
		fileCacheBaseDir = envString("FILE_CACHE_BASE_DIR", defaultCacheDir())
	)
	flag.Parse()

	initLogger(*logLevel)

	spotifyRedirectURLstr := fmt.Sprintf("%s/recommendations/v1/spotify/auth/callback", *spotifyRedirectHost)
	redirectURL, err := url.Parse(spotifyRedirectURLstr)
	if err != nil {
		slogutil.Fatal("Could not parse redirect URL", slogutil.Error(err), slog.String("spotifyRedirectSpotifyURL", spotifyRedirectURLstr))
	}

	uiRedirectURLstr := fmt.Sprintf("%s/recommendations/v1/spotify/auth/ui-redirect", *spotifyRedirectHost)
	uiRedirectURL, err := url.Parse(uiRedirectURLstr)
	if err != nil {
		slogutil.Fatal("Could not parse redirect URL", slogutil.Error(err), slog.String("spotifyRedirectSpotifyURL", spotifyRedirectURLstr))
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(50 * time.Minute)) // haha wow.

	r.Get("/status", getStatus())
	r.Method(http.MethodGet, "/metrics", promhttp.Handler())

	authAdaptor := recommendations.NewSpotifyAuthAdaptor(clientID, clientSecret, *redirectURL, *uiRedirectURL)
	r.Get(authAdaptor.Path(), authAdaptor.TokenCallbackHandler())
	r.Get(authAdaptor.UIRedirectPath(), authAdaptor.UIRedirectHandler())

	recommendatinsHandler, err := getRecommendationsHandler(authAdaptor, persistenceFactoryWith(fileCacheBaseDir))
	if err != nil {
		slogutil.Fatal("Setting up recommendations handler", slogutil.Error(err))
	}
	r.Mount("/recommendations", recommendatinsHandler)

	fs := http.FileServer(http.Dir("./static"))
	r.Handle("/*", srv.RedirectOn404(fs, "/index.html"))

	errs := make(chan error, 2)
	go func() {
		slog.Info("Starting server", slog.String("addr", *addr))
		errs <- http.ListenAndServe(*addr, r)
	}()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		<-c
		errs <- nil
	}()

	err = <-errs
	if err != nil && err != http.ErrServerClosed {
		slogutil.Fatal("Shutting down", slogutil.Error(err))
	}

	slog.Info("Server shutdown")
}

func getRecommendationsHandler(authAdaptor *recommendations.AuthAdaptor, persistedKV kvPersistenceFactory) (*chi.Mux, error) {
	serviceCache := keyvaluestore.Combine(keyvaluestore.InMemoryStore(), persistedKV("cache"))
	spotifyCache := keyvaluestore.Combine(keyvaluestore.InMemoryStore(), persistedKV("spotify-provider"))
	recommendatinsHandler := recommendations.NewRouter(
		recommendations.NewServiceFactory(serviceCache, recommendations.NewDummyUserPreferenceProvider()),
		recommendations.NewSpotifyProviderFactory(spotifyCache),
		authAdaptor,
	)
	return recommendatinsHandler, nil
}

type kvPersistenceFactory func(prefix string) keyvaluestore.KV

func persistenceFactoryWith(fileCacheBaseDir string) kvPersistenceFactory {
	return func(prefix string) keyvaluestore.KV {
		return keyvaluestore.JSONDiskStore(path.Join(fileCacheBaseDir, "recommendations", prefix))
	}
}

func envString(env, fallback string) string {
	if e := os.Getenv(env); e != "" {
		return e
	}
	return fallback
}

func defaultCacheDir() string {
	var baseDir string
	if homeDir, err := os.UserHomeDir(); err != nil {
		baseDir, _ = os.Getwd()
	} else {
		baseDir = homeDir
	}
	return path.Join(baseDir, ".recommendli")
}

func getStatus() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("OK")) // nolint
	})
}

func initLogger(logLevel string) {
	level, ok := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}[strings.ToLower(logLevel)]
	if !ok {
		level = slog.LevelInfo
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	})))
}
