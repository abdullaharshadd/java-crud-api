package smartcontact

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"migrated-app/internal/config"
	"migrated-app/internal/smartcontact/handler"
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/service"

	"github.com/gin-gonic/gin"
)

// RunApplication starts the SmartContactApplication.
func RunApplication() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	userRepository := repository.NewUserRepository(db)
	userService := service.NewUserServiceImp(userRepository)

	r := gin.Default()
	uc := handler.NewUserController(userService)
	uc.RegisterRoutes(r)

	server := &http.Server{
		Addr:           ":" + cfg.Port,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe failed: %v", err)
		}
	}()

	<-ctx.Done()

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutCancel()

	if err := server.Shutdown(shutCtx); err != nil {
		log.Fatalf("Shutdown failed: %v", err)
	}

	log.Println("Server exiting")
}