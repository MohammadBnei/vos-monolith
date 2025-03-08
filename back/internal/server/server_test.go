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

// MockLogger is a mock implementation of logger.Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug() *zerolog.Event {
	args := m.Called()
	return args.Get(0).(*zerolog.Event)
}

func (m *MockLogger) Info() *zerolog.Event {
	args := m.Called()
	return args.Get(0).(*zerolog.Event)
}

func (m *MockLogger) Warn() *zerolog.Event {
	args := m.Called()
	return args.Get(0).(*zerolog.Event)
}

func (m *MockLogger) Error() *zerolog.Event {
	args := m.Called()
	return args.Get(0).(*zerolog.Event)
}

func (m *MockLogger) Fatal() *zerolog.Event {
	args := m.Called()
	return args.Get(0).(*zerolog.Event)
}

func (m *MockLogger) Panic() *zerolog.Event {
	args := m.Called()
	return args.Get(0).(*zerolog.Event)
}

func (m *MockLogger) With() zerolog.Context {
	args := m.Called()
	return args.Get(0).(zerolog.Context)
}

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

	mockLogger := new(MockLogger)

	server := NewServer(cfg, mockLogger)

	assert.NotNil(t, server)
	assert.Equal(t, cfg, server.cfg)
	assert.Equal(t, mockLogger, server.log)
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

	mockLogger := new(MockLogger)
	mockEvent := new(MockEvent)

	// Setup expectations
	mockLogger.On("Debug").Return(mockEvent)
	mockEvent.On("Msg", "Health check request").Return()

	// Create server
	server := NewServer(cfg, mockLogger)
	server.setupRoutes()

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)

	// Create a gin context with the logger
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("logger", mockLogger)

	// Call the health check handler
	server.healthCheck(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "Test App", response["app"])

	// Verify mock expectations
	mockLogger.AssertExpectations(t)
	mockEvent.AssertExpectations(t)
}

func TestLoggerMiddleware(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		AppName:  "Test App",
		HTTPPort: "8080",
	}

	mockLogger := new(MockLogger)
	mockEvent := new(MockEvent)

	// Setup expectations for request start
	mockLogger.On("Debug").Return(mockEvent).Times(2)
	mockEvent.On("Str", "method", "GET").Return(mockEvent).Times(2)
	mockEvent.On("Str", "path", "/test").Return(mockEvent).Times(2)
	mockEvent.On("Int", "status", 200).Return(mockEvent).Once()
	mockEvent.On("Msg", "Request started").Return().Once()
	mockEvent.On("Msg", "Request completed").Return().Once()

	// Create server
	server := NewServer(cfg, mockLogger)

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
	logger, exists := c.Get("logger")
	assert.True(t, exists)
	assert.Equal(t, mockLogger, logger)

	// Verify mock expectations
	mockLogger.AssertExpectations(t)
	mockEvent.AssertExpectations(t)
}
