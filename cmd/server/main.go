package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"

	"migrated-app/internal/smartcontact/handler"
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/service"
)

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    user_id       SERIAL PRIMARY KEY,
    user_name     TEXT NOT NULL,
    user_email    TEXT NOT NULL UNIQUE,
    user_password TEXT NOT NULL,
    user_role     TEXT NOT NULL DEFAULT '',
    user_about    TEXT NOT NULL DEFAULT ''
);
`

func buildRouter() http.Handler {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Warn().Msg("DATABASE_URL not set; repository operations will fail")
	}

	var db *sql.DB
	if dsn != "" {
		var err error
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to open database connection")
		}

		// Retry ping to allow the DB container to finish starting up.
		for i := 0; i < 10; i++ {
			if err = db.Ping(); err == nil {
				break
			}
			log.Warn().Err(err).Msg("database ping failed; retrying in 1s")
			time.Sleep(time.Second)
		}
		if err != nil {
			log.Warn().Err(err).Msg("database ping still failing; continuing anyway")
		}

		// Ensure the schema exists.
		if _, err := db.Exec(createUsersTable); err != nil {
			log.Fatal().Err(err).Msg("failed to create users table")
		}
	}

	repo, err := repository.NewPostgresUserRepository(db)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create user repository")
	}

	userSvc, err := service.NewUserService(repo)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create user service")
	}

	h := handler.NewHandler(userSvc, nil)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	h.RegisterRoutes(r)

	return r
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