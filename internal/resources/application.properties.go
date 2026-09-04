package resources

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"github.com/jmoiron/sqlx"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"migrated-app/internal/smartcontact/handler"
	"migrated-app/internal/smartcontact/repository"
	"migrated-app/internal/smartcontact/service"
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
		log.Fatalf("Fatal error config file: %s \\n", err)
	}

	dbURL := viper.GetString("spring.datasource.url")
	dbUsername := viper.GetString("spring.datasource.username")
	dbPassword := viper.GetString("spring.datasource.password")

	// Update the dbURL to match PostgreSQL dialect.
	dbURL = "postgres://" + dbUsername + ":" + dbPassword + "@localhost:5432/barcode?sslmode=disable"

	db, err := sqlx.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	port := viper.GetString("server.port")

	userRepository := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepository)
	userController := handler.NewUserController(userService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/admin", func(r chi.Router) {
		r.Post("/users", userController.SaveUserHandler)
		r.Get("/users", userController.FetchUserListHandler)
		r.Get("/users/{id:[0-9]+}", userController.FetchUserByIDHandler)
		r.Put("/users/{id:[0-9]+}", userController.UpdateUserHandler)
		r.Delete("/users/{id:[0-9]+}", userController.DeleteUserHandler)
		r.Get("/users/name/{name}", userController.FindUserByNameHandler)
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
			log.Fatalf("listen: %s\\\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	delayedCancel := time.AfterFunc(5*time.Second, cancel)
	delayedCancel.Stop()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
	return nil
}
