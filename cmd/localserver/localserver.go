package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/kristofferostlund/recommendli/internal/recommendations"
	"github.com/kristofferostlund/recommendli/pkg/keyvaluestore"
	"github.com/kristofferostlund/recommendli/pkg/logging"
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

		clientID     = envString("SPOTIFY_ID", "")
		clientSecret = envString("SPOTIFY_SECRET", "")
	)
	flag.Parse()

	log := logging.New(logging.LevelDebug, logging.FormatConsolePretty)
	// nolint errcheck
	defer log.Sync()

	spotifyRedirectURLstr := fmt.Sprintf("%s/recommendations/v1/spotify/auth/callback", *spotifyRedirectHost)
	redirectURL, err := url.Parse(spotifyRedirectURLstr)
	if err != nil {
		log.Fatal("Could not parse redirect URL", err, "spotifyRedirectSpotifyURL", spotifyRedirectURLstr)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/status", getStatus())
	r.Method(http.MethodGet, "/metrics", promhttp.Handler())

	authAdaptor := recommendations.NewSpotifyAuthAdaptor(clientID, clientSecret, *redirectURL, log)
	r.Get(authAdaptor.Path(), authAdaptor.TokenCallbackHandler())
	r.Mount("/recommendations", recommendations.NewRouter(
		recommendations.NewServiceFactory(log),
		recommendations.NewSpotifyProviderFactory(log, keyvaluestore.JSONDiskStore("/Users/kristofferostlund/.recommendli")),
		authAdaptor,
		log,
	))

	errs := make(chan error, 2)
	go func() {
		log.Info("Starting server", "addr", *addr)
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
		log.Fatal("Shutting down", err)
	}

	log.Info("Server shutdown")
}

func envString(env, fallback string) string {
	if e := os.Getenv(env); e != "" {
		return e
	}
	return fallback
}

func getStatus() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("OK"))
	})
}
