package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kelseyhightower/envconfig"
	"github.com/kristofferostlund/recommendli/internal/recommendations"
	"github.com/kristofferostlund/recommendli/internal/sqlite"
	"github.com/kristofferostlund/recommendli/pkg/migrations"
	"github.com/kristofferostlund/recommendli/pkg/singleflight"
	"github.com/kristofferostlund/recommendli/pkg/slogutil"
	"github.com/kristofferostlund/recommendli/pkg/srv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	SpotifyClientID     string `envconfig:"SPOTIFY_ID"`
	SpotifyClientSecret string `envconfig:"SPOTIFY_SECRET"`
	SpotifyRedirectHost string `envconfig:"SPOTIFY_REDIRECT_HOST" default:"http://127.0.0.1:9999"`
	LogLevel            string `envconfig:"LOG_LEVEL" default:"info"`
	Addr                string `envconfig:"ADDR"`
	Port                string `envconfig:"PORT" default:"9999"`
	FileCacheBaseDir    string `envconfig:"FILE_CACHE_BASE_DIR" default:"/tmp/recommendli"`
	SQLiteDBPath        string `envconfig:"SQLITE_DB_PATH" default:"/tmp/recommendli.sqlite"`
}

var migrationsDir = fmt.Sprintf("file://%s", absolutePathTo("./migrations"))

func main() {
	cfg, err := loadConfig()
	if err != nil {
		slogutil.Fatal("Could not load config", slogutil.Error(err))
	}

	slogutil.InitDefaultLogger(cfg.LogLevel)

	if err := migrations.Up(migrationsDir, fmt.Sprintf("sqlite3://%s", cfg.SQLiteDBPath)); err != nil {
		slogutil.Fatal("Running migrations", slogutil.Error(err))
	}

	sqliteDB, err := sqlite.Open(cfg.SQLiteDBPath)
	if err != nil {
		slogutil.Fatal("Could not open sqlite database", slogutil.Error(err))
	}
	defer sqliteDB.Close()
	db := sqlite.Wrap(sqliteDB)

	spotifyRedirectURLstr := fmt.Sprintf("%s/recommendations/v1/spotify/auth/callback", cfg.SpotifyRedirectHost)
	redirectURL, err := url.Parse(spotifyRedirectURLstr)
	if err != nil {
		slogutil.Fatal("Could not parse redirect URL", slogutil.Error(err), slog.String("spotifyRedirectSpotifyURL", spotifyRedirectURLstr))
	}

	uiRedirectURLstr := fmt.Sprintf("%s/recommendations/v1/spotify/auth/ui-redirect", cfg.SpotifyRedirectHost)
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

	authAdaptor := recommendations.NewSpotifyAuthAdaptor(cfg.SpotifyClientID, cfg.SpotifyClientSecret, *redirectURL, *uiRedirectURL)
	r.Get(authAdaptor.Path(), authAdaptor.TokenCallbackHandler())
	r.Get(authAdaptor.UIRedirectPath(), authAdaptor.UIRedirectHandler())

	recommendatinsHandler, err := getRecommendationsHandler(authAdaptor, sqlitePeristenceFactory(db), sqlite.NewTrackIndex(db, recommendations.TrackKey), sqlite.NewLocker(db))
	if err != nil {
		slogutil.Fatal("Setting up recommendations handler", slogutil.Error(err))
	}
	r.Mount("/recommendations", recommendatinsHandler)

	staticDir := "./static/dist"
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		staticDir = "./static"
	}
	fs := http.FileServer(http.Dir(staticDir))
	r.Handle("/*", srv.RedirectOn404(fs, "/index.html"))

	server := &http.Server{Addr: cfg.Addr, Handler: r}
	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("Starting server", slog.String("addr", cfg.Addr))
		serverErrors <- server.ListenAndServe()
	}()

	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(shutdownSignals, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err = <-serverErrors:
	case sig := <-shutdownSignals:
		slog.Info("Shutdown signal received", slog.String("signal", sig.String()))
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = server.Shutdown(shutdownCtx)
		if err == nil {
			err = <-serverErrors
		}
	}

	if err != nil && err != http.ErrServerClosed {
		slogutil.Fatal("Shutting down", slogutil.Error(err))
	}

	slog.Info("Server shutdown")
}

func getRecommendationsHandler(authAdaptor *recommendations.AuthAdaptor, persistedKV kvPersistenceFactory, trackIndex recommendations.TrackIndex, sfLocker singleflight.Locker) (*chi.Mux, error) {
	serviceCache := persistedKV("cache")
	spotifyCache := persistedKV("spotify-provider")

	recommendatinsHandler := recommendations.NewRouter(
		recommendations.NewServiceFactory(serviceCache, recommendations.NewDummyUserPreferenceProvider(), trackIndex, sfLocker),
		recommendations.NewSpotifyProviderFactory(spotifyCache),
		authAdaptor,
	)
	return recommendatinsHandler, nil
}

type kvPersistenceFactory func(prefix string) recommendations.KeyValueStore

func sqlitePeristenceFactory(db *sqlite.DB) kvPersistenceFactory {
	return func(kind string) recommendations.KeyValueStore {
		return sqlite.NewKeyValueStore(db, kind)
	}
}

func getStatus() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("OK")) // nolint
	})
}

func loadConfig() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return cfg, fmt.Errorf("loading config: %w", err)
	}

	// Railway supplies PORT at runtime. ADDR remains available as an explicit
	// override for local development and other deployment environments.
	if cfg.Addr == "" {
		cfg.Addr = "0.0.0.0:" + cfg.Port
	}

	return cfg, nil
}

func absolutePathTo(relative string) string {
	absolute, err := filepath.Abs(relative)
	if err != nil {
		return filepath.Clean(relative)
	}
	return absolute
}
