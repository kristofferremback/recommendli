package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/kristofferostlund/recommendli/pkg/logging"
	"github.com/kristofferostlund/recommendli/pkg/spotify"
	"github.com/kristofferostlund/recommendli/server"
)

const (
	defaultAddr = ":9999"
)

func main() {
	var (
		addr = flag.String("addr", defaultAddr, "HTTP address")
		// ctx  = context.Background()

		clientID     = envString("SPOTIFY_ID", "")
		clientSecret = envString("SPOTIFY_SECRET", "")
	)
	flag.Parse()

	log := logging.New(logging.LevelInfo, logging.FormatConsolePretty)

	redirectURL, err := url.Parse(fmt.Sprintf("http://localhost%s/v1/spotify/auth/callback", *addr))
	if err != nil {
		log.Fatal("Could not parse redirect URL", err)
	}

	authSvc := spotify.NewAuthService(clientID, clientSecret, *redirectURL, log)
	srv := server.New(authSvc, log.With("addr", addr))

	errs := make(chan error, 2)
	go func() {
		log.With("addr", *addr).Info("Starting server")
		errs <- http.ListenAndServe(*addr, srv)
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
