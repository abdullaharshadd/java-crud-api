package main

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"

	"migrated-app/internal/config"
	"migrated-app/internal/httpapi"
	"migrated-app/internal/user"
)

// mysqlDSN converts DATABASE_URL into a go-sql-driver/mysql DSN. A
// "mysql://user:pass@host:port/db?params" URL is translated; any other value
// is assumed to already be a native driver DSN and is returned unchanged.
func mysqlDSN(raw string) string {
	if !strings.HasPrefix(raw, "mysql://") {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	cfg := mysql.NewConfig()
	cfg.Net = "tcp"
	cfg.Addr = u.Host
	cfg.DBName = strings.TrimPrefix(u.Path, "/")
	if u.User != nil {
		cfg.User = u.User.Username()
		if p, ok := u.User.Password(); ok {
			cfg.Passwd = p
		}
	}
	params := map[string]string{}
	for k, v := range u.Query() {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	if len(params) > 0 {
		cfg.Params = params
	}
	cfg.ParseTime = true
	return cfg.FormatDSN()
}

func buildRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	if cfg.DatabaseURL == "" {
		log.Fatal().Msg("DATABASE_URL is not set")
	}

	db, err := sql.Open("mysql", mysqlDSN(cfg.DatabaseURL))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open database")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := user.EnsureSchema(ctx, db); err != nil {
		log.Error().Err(err).Msg("failed to ensure user schema")
	}

	repo := user.NewMySQLRepository(db)
	svc := user.NewService(repo)
	h := httpapi.NewHandler(svc)

	r.Mount("/", httpapi.Recoverer(h.Routes()))

	return r
}