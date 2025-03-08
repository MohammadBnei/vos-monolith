// Package logger provides a standardized logging interface for the application.
package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// Environment types
const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
	EnvTest        = "test"
)

// Valid log levels
const (
	LevelTrace = "trace"
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
	LevelFatal = "fatal"
	LevelPanic = "panic"
)

func init() {
	// Set default time format to ISO8601
	zerolog.TimeFieldFormat = time.RFC3339
	
	// Always enable stack traces for errors
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
}

// Config holds logger configuration options
type Config struct {
	// Environment determines the output format (development/production)
	Environment string
	// Level sets the minimum log level
	Level string
	// Output is where the logs will be written
	Output io.Writer
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	// Determine environment from ENV variable
	env := EnvDevelopment
	if os.Getenv("ENV") == "production" {
		env = EnvProduction
	}

	return Config{
		Environment: env,
		Level:       LevelInfo,
		Output:      os.Stdout,
	}
}

// New returns a new zerolog.Logger instance with the given log level.
// The log level should be one of the following: trace, debug, info, warn, error, fatal, panic.
// If an invalid log level is provided, the logger defaults to info.
func New(level string) zerolog.Logger {
	return NewWithConfig(Config{
		Environment: EnvDevelopment,
		Level:       level,
		Output:      os.Stdout,
	})
}

// NewWithConfig creates a logger with the specified configuration
func NewWithConfig(config Config) zerolog.Logger {
	logLevel, err := zerolog.ParseLevel(config.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid log level '%s', defaulting to 'info'\n", config.Level)
		logLevel = zerolog.InfoLevel
	}

	var output io.Writer
	if config.Output == nil {
		config.Output = os.Stdout
	}

	// Configure output format based on environment
	if config.Environment == EnvProduction {
		// In production, use JSON format
		output = config.Output
	} else {
		// In development/test, use pretty console output
		output = zerolog.ConsoleWriter{
			Out:        config.Output,
			TimeFormat: time.RFC3339,
			NoColor:    os.Getenv("NO_COLOR") != "",
		}
	}

	// Create and configure the logger
	return zerolog.New(output).
		Level(logLevel).
		With().
		Timestamp().
		Caller().
		Logger()
}
