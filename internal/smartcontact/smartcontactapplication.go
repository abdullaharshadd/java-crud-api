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
)

// buildRouter sets up the HTTP router with all necessary routes and middleware.
func buildRouter(db *sqlx.DB) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	userRepository := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepository)
	userController := handler.NewUserController(userService)

	r.Route("/admin", func(r chi.Router) {
		r.Post("/users", userController.SaveUserHandler)
		r.Get("/users", userController.FetchUserListHandler)
		r.Get("/users/{id}", userController.FetchUserByIDHandler)
		r.Delete("/users/{id}", userController.DeleteUserHandler)
		r.Put("/users/{id}", userController.UpdateUserHandler)
		r.Get("/users/name/{name}", userController.FindUserByNameHandler)
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return r
}

// initDatabase initializes the database connection.
func initDatabase() (*sqlx.DB, error) {
	dsn := "postgres://user:password@localhost/dbname?sslmode=disable"
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping the database: %w", err)
	}

	// Ensure the schema is created
	if err := createUserTable(db); err != nil {
		return nil, fmt.Errorf("failed to create user table: %w", err)
	}

	return db, nil
}

// createUserTable creates the user table if it does not exist.
func createUserTable(db *sqlx.DB) error {
	query := `CREATE TABLE IF NOT EXISTS "user" (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	_, err := db.Exec(query)
	return err
}

// startServer starts the HTTP server.
func startServer(router http.Handler) error {
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	shutdownCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

defer cancel()

	<-shutdownCtx.Done()
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("server gracefully stopped")
	return nil
}
