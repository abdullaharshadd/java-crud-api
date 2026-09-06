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
	"migrated-app/internal/config"
	"migrated-app/internal/smartcontact/handler"
	"migrated-app/internal/smartcontact/service"
	"migrated-app/internal/smartcontact/repository"
)

// buildRouter builds the HTTP router for the application.
func buildRouter(us service.UserService) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	uc := handler.NewUserController(us)
	uc.RegisterRoutes(r)

	r.Route("/admin", func(r chi.Router) {
		// Admin-specific routes can be registered here
	})

	return r
}

// RunApplication starts the SmartContactApplication.
func RunApplication() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	dbURL := cfg.DatabaseURL
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	userRepository := repository.NewUserRepository(db)
	userService := service.NewUserServiceImp(userRepository)

	router := buildRouter(userService)

	server := &http.Server{
		Addr:           ":" + cfg.Port,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe failed: %v", err)
		}
	}()

	<-ctx.Done()

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown failed: %v", err)
	}

	log.Println("Server exiting")
}