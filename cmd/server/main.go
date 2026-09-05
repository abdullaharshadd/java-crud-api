package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"migrated-app/internal/config"
	"migrated-app/internal/smartcontact/handler"
	"migrated-app/internal/smartcontact/service"
	"migrated-app/internal/smartcontact/repository"
	"database/sql"
	_ "github.com/lib/pq"
)

func buildRouter() *gin.Engine {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	dbRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(dbRepo)
	userCtrl := handler.NewUserController(userSvc)
	handler.RegisterRoutes(gin.Default(), userCtrl)
	return gin.Default()
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: buildRouter(),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	log.Info().Msg("server started on :" + cfg.Port)
	<-ctx.Done()

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	delayedCancel := time.AfterFunc(10*time.Second, cancel)
	delayedCancel.Stop()

	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown failed")
	}

	log.Info().Msg("Server exiting")
}