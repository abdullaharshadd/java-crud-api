package config

import "github.com/spf13/viper"

type Config struct {
	DatabaseURL string `mapstructure:"DATABASE_URL"`
	Port        string `mapstructure:"PORT"`
	JWTSecret   string `mapstructure:"JWT_SECRET"`
}

// Load reads configuration from the environment. viper.AutomaticEnv() alone
// only overlays env vars onto keys viper already knows about (via SetDefault
// or BindEnv) at Unmarshal time — DATABASE_URL and JWT_SECRET had neither, so
// they silently unmarshalled as empty strings regardless of what was actually
// set in the environment (PORT happened to work only because of its
// SetDefault call). Each key viper.Unmarshal is expected to populate from the
// environment must be explicitly bound.
func Load() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetDefault("PORT", "8080")
	if err := viper.BindEnv("DATABASE_URL"); err != nil {
		return nil, err
	}
	if err := viper.BindEnv("JWT_SECRET"); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
