package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"

	smartcontact "migrated-app/internal/smartcontact"
	"migrated-app/internal/smartcontact/repository"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://app:app@migrator-sandbox-db:5432/app?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var db *sql.DB
	var err error

	// Retry connecting to the database up to 10 times with backoff.
	for i := 0; i < 10; i++ {
		db, err = sql.Open("pgx", databaseURL)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			err = db.PingContext(ctx)
			cancel()
		}
		if err == nil {
			break
		}
		log.Warn().Err(err).Msgf("db not ready, retry %d/10", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal().Err(err).Msg("could not connect to database")
	}

	// Ensure schema exists.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := repository.InitSchema(ctx, db); err != nil {
		cancel()
		log.Fatal().Err(err).Msg("failed to initialise schema")
	}
	cancel()

	router := smartcontact.BuildRouterWith(db)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Info().Msgf("listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server shutdown error")
	}
	log.Info().Msg("server stopped")
}