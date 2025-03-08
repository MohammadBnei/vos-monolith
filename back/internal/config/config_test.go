package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test default values
	t.Run("default values", func(t *testing.T) {
		// Clear any existing env vars that might affect the test
		os.Unsetenv("APP_NAME")
		os.Unsetenv("HTTP_PORT")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("REDIS_URL")
		os.Unsetenv("JWT_SECRET")

		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if cfg.AppName != "Voc on Steroid" {
			t.Errorf("Expected AppName to be 'Voc on Steroid', got %s", cfg.AppName)
		}
		if cfg.HTTPPort != "8080" {
			t.Errorf("Expected HTTPPort to be '8080', got %s", cfg.HTTPPort)
		}
		if cfg.LogLevel != "info" {
			t.Errorf("Expected LogLevel to be 'info', got %s", cfg.LogLevel)
		}
		if cfg.DatabaseURL != "postgres://user:password@localhost:5432/voconsteroid" {
			t.Errorf("Expected DatabaseURL to be 'postgres://user:password@localhost:5432/voconsteroid', got %s", cfg.DatabaseURL)
		}
		if cfg.RedisURL != "redis://localhost:6379" {
			t.Errorf("Expected RedisURL to be 'redis://localhost:6379', got %s", cfg.RedisURL)
		}
		if cfg.JWTSecret != "secret" {
			t.Errorf("Expected JWTSecret to be 'secret', got %s", cfg.JWTSecret)
		}
	})

	// Test custom values
	t.Run("custom values", func(t *testing.T) {
		// Set custom env vars
		os.Setenv("APP_NAME", "Test App")
		os.Setenv("HTTP_PORT", "9090")
		os.Setenv("LOG_LEVEL", "info")
		os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
		os.Setenv("REDIS_URL", "redis://localhost:6380")
		os.Setenv("JWT_SECRET", "test-secret")

		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if cfg.AppName != "Test App" {
			t.Errorf("Expected AppName to be 'Test App', got %s", cfg.AppName)
		}
		if cfg.HTTPPort != "9090" {
			t.Errorf("Expected HTTPPort to be '9090', got %s", cfg.HTTPPort)
		}
		if cfg.LogLevel != "info" {
			t.Errorf("Expected LogLevel to be 'info', got %s", cfg.LogLevel)
		}
		if cfg.DatabaseURL != "postgres://test:test@localhost:5432/test" {
			t.Errorf("Expected DatabaseURL to be 'postgres://test:test@localhost:5432/test', got %s", cfg.DatabaseURL)
		}
		if cfg.RedisURL != "redis://localhost:6380" {
			t.Errorf("Expected RedisURL to be 'redis://localhost:6380', got %s", cfg.RedisURL)
		}
		if cfg.JWTSecret != "test-secret" {
			t.Errorf("Expected JWTSecret to be 'test-secret', got %s", cfg.JWTSecret)
		}

		// Clean up
		os.Unsetenv("APP_NAME")
		os.Unsetenv("HTTP_PORT")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("REDIS_URL")
		os.Unsetenv("JWT_SECRET")
	})
}
