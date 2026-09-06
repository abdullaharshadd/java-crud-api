package resources

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"migrated-app/internal/smartcontact/handler"
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/service"

	_ "github.com/lib/pq"
)

// LoadConfig loads the application configuration.
func LoadConfig() error {
	viper.SetConfigName("application")
	viper.SetConfigType("properties")
	viper.AddConfigPath("./")
	viper.AddConfigPath("/etc/app/")
	viper.AddConfigPath("$HOME/.app/")
	viper.AutomaticEnv()

	return viper.ReadInConfig()
}

// BuildApplication builds the application and starts the server.
func BuildApplication() error {
	if err := LoadConfig(); err != nil {
		log.Printf("Config file not found, using environment variables: %s\n", err)
	}

	dbURL := viper.GetString("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://app:app@migrator-sandbox-db:5432/app?sslmode=disable"
	}

	db, err := sqlx.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	port := viper.GetString("PORT")
	if port == "" {
		port = "8080"
	}

	userRepository := repository.NewUserRepository(db.DB)
	userService := service.NewUserService(userRepository)
	userController := handler.NewUserController(userService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/admin", func(r chi.Router) {
		_ = userController
	})

	server := &http.Server{
		Addr:           ":" + port,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
	return nil
}