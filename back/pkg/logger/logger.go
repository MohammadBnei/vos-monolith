package logger

import (
	"os"

	"github.com/rs/zerolog"
)

// New returns a new zerolog.Logger instance with the given log level.
// The log level should be one of the following: trace, debug, info, warn, error, fatal, panic.
// If an invalid log level is provided, the logger defaults to info.
func New(level string) zerolog.Logger {
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	output := zerolog.ConsoleWriter{Out: os.Stdout}
	return zerolog.New(output).
		Level(logLevel).
		With().
		Timestamp().
		Logger()
}
