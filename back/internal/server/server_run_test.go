package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"voconsteroid/internal/config"
	"voconsteroid/internal/domain/word"
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
	
	// Create mock word service
	wordService := new(MockWordService)
	
	// Mock the GetRecentWords method which might be called during startup
	wordService.On("GetRecentWords", mock.Anything, mock.Anything, mock.Anything).
		Return([]*word.Word{}, nil)
	
	server := NewServer(cfg, log, wordService)
	
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

func TestServerWithWordAPI(t *testing.T) {
	// Skip in short mode as this test starts an actual server
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Setup
	cfg := &config.Config{
		AppName:  "Test App",
		HTTPPort: "8082", // Use a different port for testing
	}
	
	log := zerolog.New(zerolog.NewConsoleWriter()).
		Level(zerolog.DebugLevel).
		With().
		Timestamp().
		Logger()
	
	// Create mock word service with specific test data
	wordService := new(MockWordService)
	
	// Mock the GetRecentWords method for the /api/v1/words/recent endpoint
	testWords := []*word.Word{
		{
			Text:     "example",
			Language: "en",
			Definitions: []word.Definition{
				{
					Text:     "a representative form or pattern",
					WordType: "noun",
				},
			},
		},
	}
	wordService.On("GetRecentWords", mock.Anything, "en", 10).Return(testWords, nil)
	
	// Create and start server
	server := NewServer(cfg, log, wordService)
	err := server.Run()
	assert.NoError(t, err)
	
	// Give the server time to start
	time.Sleep(100 * time.Millisecond)
	
	// Test the recent words API endpoint
	resp, err := http.Get("http://localhost:8082/api/v1/words/recent?language=en")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	// Shutdown the server
	err = server.Shutdown()
	assert.NoError(t, err)
	
	// Verify our mock was called
	wordService.AssertExpectations(t)
}
