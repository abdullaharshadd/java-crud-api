package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"

	"migrated-app/internal/config"
	"migrated-app/internal/smartcontact/handler"
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/service"
)

// schemaSQL creates the users table if it doesn't already exist.
//
// MIGRATION_NOTE: The Java source used
// spring.jpa.hibernate.ddl-auto=update, which auto-created/updated the
// schema from the JPA entity at boot. Go's sqlx has no ORM-driven schema
// sync, so this replaces that auto-DDL with an explicit, idempotent
// CREATE TABLE IF NOT EXISTS matching model.User's db tags exactly
// (user_id/user_name/user_email/user_password/user_role/user_about).
const schemaSQL = `
CREATE TABLE IF NOT EXISTS users (
	user_id        SERIAL PRIMARY KEY,
	user_name      VARCHAR(255) NOT NULL,
	user_email     VARCHAR(255) NOT NULL,
	user_password  VARCHAR(255) NOT NULL,
	user_role      VARCHAR(255) NOT NULL,
	user_about     VARCHAR(500) NOT NULL DEFAULT ''
)`

// buildRouter wires the full HTTP handler: chi router with base middleware,
// the composition root's own routes (health check), and every migrated
// controller's real handlers.
//
// MIGRATION_NOTE: This replaces Spring's @ComponentScan/@EnableAutoConfiguration
// component wiring. Each layer (repository -> service -> handler) is
// constructed explicitly here rather than discovered via reflection.
func buildRouter(db *sqlx.DB) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok\n"))
	})

	userDao := repository.NewUserDao(db)
	userService := service.NewUserService(userDao)
	userController := handler.NewUserController(userService, nil)
	userController.RegisterRoutes(r)

	return r
}

func connectWithRetry(dsn string, maxRetries int, delay time.Duration) (*sqlx.DB, error) {
	var db *sqlx.DB
	var err error
	for i := 0; i < maxRetries; i++ {
		db, err = sqlx.Connect("postgres", dsn)
		if err == nil {
			return db, nil
		}
		log.Warn().Err(err).Msgf("failed to connect to database, retrying in %s (%d/%d)", delay, i+1, maxRetries)
		time.Sleep(delay)
	}
	return nil, err
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	db, err := connectWithRetry(cfg.DatabaseURL, 10, 3*time.Second)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	if _, err := db.Exec(schemaSQL); err != nil {
		log.Fatal().Err(err).Msg("failed to apply schema")
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: buildRouter(db),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	log.Info().Msgf("server started on :%s", cfg.Port)
	<-ctx.Done()

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown failed")
	}
}