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

	"migrated-app/internal/smartcontact/handler"
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/service"

	smartcontact "migrated-app/internal/smartcontact"
)

func buildRouter() http.Handler {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Warn().Msg("DATABASE_URL not set; database-backed routes will be unavailable")
		return smartcontact.BuildRouterWith(nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Error().Err(err).Msg("failed to open database connection; routes will be stubs")
		return smartcontact.BuildRouterWith(nil)
	}

	userDAO, err := repository.NewUserDao(ctx, db)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialise UserDao; routes will be stubs")
		return smartcontact.BuildRouterWith(nil)
	}

	userSvc := service.NewUserService(userDAO)
	userCtrl := handler.NewUserController(userSvc, nil)

	return smartcontact.BuildRouterWith(userCtrl)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: buildRouter(),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	log.Info().Msg("server started on :8080")
	<-ctx.Done()

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown failed")
	}
}
