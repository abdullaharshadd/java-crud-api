package smartcontact

import (
	"net/http"
	"migrated-app/internal/config"
)

// LoadConfig loads the configuration.
func LoadConfig() (*config.Config, error) {
	return config.Load()
}

// BuildRouter builds the HTTP router.
func BuildRouter(cfg *config.Config) http.Handler {
	router := http.NewServeMux()
	// TODO: Add your route definitions here.
	return router
}