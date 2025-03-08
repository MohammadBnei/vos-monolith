// Package logger provides a standardized logging interface for the application.
package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
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
}

// New returns a new zerolog.Logger instance with the given log level.
// The log level should be one of the following: trace, debug, info, warn, error, fatal, panic.
// If an invalid log level is provided, the logger defaults to info.
func New(level string) zerolog.Logger {
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid log level '%s', defaulting to 'info'\n", level)
		logLevel = zerolog.InfoLevel
	}

	// Enable stack traces for errors in debug and error levels
	if logLevel == zerolog.ErrorLevel || logLevel == zerolog.DebugLevel {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	}

	// Configure console output with colors and readable format
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    os.Getenv("NO_COLOR") != "",
	}

	// Create and configure the logger
	return zerolog.New(output).
		Level(logLevel).
		With().
		Timestamp().
		Caller().
		Logger()
}
