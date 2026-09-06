package smartcontact

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	"migrated-app/internal/smartcontact/handler"
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/service"
	"migrated-app/internal/resources"
)

// SmartContactApplicationTests contains the test cases for the SmartContactApplication.
func SmartContactApplicationTests() {
	if err := resources.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := sqlx.Connect("postgres", viper.GetString("database.url"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute * 5)

	buildRouter := buildRouter(db)
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-shutdown
		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		db.Close()
		cancel()
		log.Println("Server gracefully shut down.")
	}()

	server := &http.Server{
		Addr: ":8080",
		Handler: buildRouter,
	}

	log.Printf("Starting server on port %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}

	<-ctx.Done()
}