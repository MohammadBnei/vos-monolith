package config

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	AppName     string `env:"APP_NAME" envDefault:"Voc on Steroid"`
	HTTPPort    string `env:"HTTP_PORT" envDefault:"8080"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
	DatabaseURL string `env:"DATABASE_URL" envDefault:"postgres://user:password@localhost:5432/voconsteroid"`
	RedisURL    string `env:"REDIS_URL" envDefault:"redis://localhost:6379"`
	JWTSecret   string `env:"JWT_SECRET" envDefault:"secret"`
}

// LoadConfig loads configuration from environment variables into a Config object.
// Environment variables are matched by exact name, and default values are used
// if the corresponding environment variable is not defined.
//
func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
