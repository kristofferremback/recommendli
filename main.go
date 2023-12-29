package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kelseyhightower/envconfig"
	"github.com/kristofferostlund/recommendli/internal/recommendations"
	"github.com/kristofferostlund/recommendli/pkg/keyvaluestore"
	"github.com/kristofferostlund/recommendli/pkg/slogutil"
	"github.com/kristofferostlund/recommendli/pkg/srv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	SpotifyClientID     string `envconfig:"SPOTIFY_ID"`
	SpotifyClientSecret string `envconfig:"SPOTIFY_SECRET"`
	SpotifyRedirectHost string `envconfig:"SPOTIFY_REDIRECT_HOST" default:"http://0.0.0.0:9999"`
	LogLevel            string `envconfig:"LOG_LEVEL" default:"info"`
	Addr                string `envconfig:"ADDR" default:"0.0.0.0:9999"`
	FileCacheBaseDir    string `envconfig:"FILE_CACHE_BASE_DIR" default:"/tmp/recommendli"`
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		slogutil.Fatal("Could not load config", slogutil.Error(err))
	}

	slogutil.InitDefaultLogger(cfg.LogLevel)

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

	recommendatinsHandler, err := getRecommendationsHandler(authAdaptor, persistenceFactoryWith(cfg.FileCacheBaseDir))
	if err != nil {
		slogutil.Fatal("Setting up recommendations handler", slogutil.Error(err))
	}
	r.Mount("/recommendations", recommendatinsHandler)

	fs := http.FileServer(http.Dir("./static"))
	r.Handle("/*", srv.RedirectOn404(fs, "/index.html"))

	errs := make(chan error, 2)
	go func() {
		slog.Info("Starting server", slog.String("addr", cfg.Addr))
		errs <- http.ListenAndServe(cfg.Addr, r)
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

func getStatus() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("OK")) // nolint
	})
}

func loadConfig() (Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return cfg, fmt.Errorf("loading config: %w", err)
	}
	return cfg, nil
}
