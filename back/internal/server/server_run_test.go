package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"voconsteroid/internal/config"
)

func TestServerRunAndShutdown(t *testing.T) {
	// Skip in short mode as this test starts an actual server
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Setup
	cfg := &config.Config{
		AppName:  "Test App",
		HTTPPort: "8081", // Use a different port for testing
	}
	
	// Use a real logger for this test
	log := &zerologLogger{
		zerolog.New(zerolog.NewTestWriter(t)).Level(zerolog.DebugLevel),
	}
	
	server := NewServer(cfg, log)
	
	// Start the server
	err := server.Run()
	assert.NoError(t, err)
	
	// Give the server time to start
	time.Sleep(100 * time.Millisecond)
	
	// Make a request to the server
	resp, err := http.Get("http://localhost:8081/health")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	// Shutdown the server
	err = server.Shutdown()
	assert.NoError(t, err)
}

// zerologLogger is a simple implementation of logger.Logger for testing
type zerologLogger struct {
	zerolog.Logger
}

func (l *zerologLogger) Debug() *zerolog.Event {
	return l.Logger.Debug()
}

func (l *zerologLogger) Info() *zerolog.Event {
	return l.Logger.Info()
}

func (l *zerologLogger) Warn() *zerolog.Event {
	return l.Logger.Warn()
}

func (l *zerologLogger) Error() *zerolog.Event {
	return l.Logger.Error()
}

func (l *zerologLogger) Fatal() *zerolog.Event {
	return l.Logger.Fatal()
}

func (l *zerologLogger) Panic() *zerolog.Event {
	return l.Logger.Panic()
}

func (l *zerologLogger) With() zerolog.Context {
	return l.Logger.With()
}
