package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"migrated-app/internal/smartcontact"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	delayedStop := make(chan struct{})
	defer close(delayedStop)
	defer stop()

	config, err := smartcontact.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	srv := &http.Server{
		Addr:    ":" + config.Port,
		Handler: smartcontact.BuildRouter(config),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	log.Info().Msg("server started on :" + config.Port)
	<-ctx.Done()

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	postShut := make(chan struct{})
	delayedShutdown(shutCtx, srv, postShut, cancel)
	<-postShut
}

func delayedShutdown(shutCtx context.Context, srv *http.Server, postShut chan struct{}, cancel context.CancelFunc) {
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown failed")
	}
	cancel()
	close(postShut)
}