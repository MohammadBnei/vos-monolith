package logger

import (
	"os"

	"github.com/rs/zerolog"
)

type Logger interface {
	Debug() *zerolog.Event
	Info() *zerolog.Event
	Warn() *zerolog.Event
	Error() *zerolog.Event
	Fatal() *zerolog.Event
	Panic() *zerolog.Event
	With() zerolog.Context
}

type zerologLogger struct {
	zerolog.Logger
}

// New returns a new Logger instance with the given log level.
// The log level should be one of the following: trace, debug, info, warn, error, fatal, panic.
// If an invalid log level is provided, the logger defaults to info.
func New(level string) Logger {
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	output := zerolog.ConsoleWriter{Out: os.Stdout}
	return &zerologLogger{
		zerolog.New(output).
			Level(logLevel).
			With().
			Timestamp().
			Logger(),
	}
}
