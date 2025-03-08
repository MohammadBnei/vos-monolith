package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"voconsteroid/internal/config"
)

// MockEvent is a mock implementation of zerolog.Event
type MockEvent struct {
	mock.Mock
}

func (m *MockEvent) Str(key, val string) *zerolog.Event {
	args := m.Called(key, val)
	return args.Get(0).(*zerolog.Event)
}

func (m *MockEvent) Int(key string, val int) *zerolog.Event {
	args := m.Called(key, val)
	return args.Get(0).(*zerolog.Event)
}

func (m *MockEvent) Msg(msg string) {
	m.Called(msg)
}

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		AppName:  "Test App",
		HTTPPort: "8080",
	}

	logger := zerolog.New(zerolog.NewTestWriter())

	server := NewServer(cfg, logger)

	assert.NotNil(t, server)
	assert.Equal(t, cfg, server.cfg)
	assert.Equal(t, logger, server.log)
	assert.NotNil(t, server.router)
	assert.NotNil(t, server.serverErrors)
	assert.NotNil(t, server.shutdown)
}

func TestHealthCheck(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		AppName:  "Test App",
		HTTPPort: "8080",
	}

	// Create a test writer to capture log output
	testWriter := zerolog.NewTestWriter()
	logger := zerolog.New(testWriter)

	// Create server
	server := NewServer(cfg, logger)
	server.setupRoutes()

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)

	// Create a gin context with the logger
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("logger", logger)

	// Call the health check handler
	server.healthCheck(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "Test App", response["app"])

	// Verify log output contains our message
	assert.Contains(t, testWriter.String(), "Health check request")
}

func TestLoggerMiddleware(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		AppName:  "Test App",
		HTTPPort: "8080",
	}

	// Create a test writer to capture log output
	testWriter := zerolog.NewTestWriter()
	logger := zerolog.New(testWriter)

	// Create server
	server := NewServer(cfg, logger)

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	// Create a gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Writer.WriteHeader(http.StatusOK)

	// Call the middleware
	middleware := server.loggerMiddleware()
	middleware(c)

	// Assert that logger was added to context
	contextLogger, exists := c.Get("logger")
	assert.True(t, exists)
	assert.Equal(t, logger, contextLogger)

	// Verify log output contains our messages
	logOutput := testWriter.String()
	assert.Contains(t, logOutput, "Request started")
	assert.Contains(t, logOutput, "Request completed")
	assert.Contains(t, logOutput, "GET")
	assert.Contains(t, logOutput, "/test")
}
