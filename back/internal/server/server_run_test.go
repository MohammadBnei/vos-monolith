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
	log := zerolog.New(zerolog.NewConsoleWriter()).
		Level(zerolog.DebugLevel).
		With().
		Timestamp().
		Logger()
	
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
