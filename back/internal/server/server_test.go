package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"voconsteroid/internal/config"
	"voconsteroid/internal/domain/word"
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

// MockWordService is a mock implementation of word.Service
type MockWordService struct {
	mock.Mock
}

func (m *MockWordService) Search(ctx context.Context, text, language string) (*word.Word, error) {
	args := m.Called(ctx, text, language)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*word.Word), args.Error(1)
}

func (m *MockWordService) GetRelatedWords(ctx context.Context, wordID string) (*word.RelatedWords, error) {
	args := m.Called(ctx, wordID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*word.RelatedWords), args.Error(1)
}

func (m *MockWordService) GetRecentWords(ctx context.Context, language string, limit int) ([]*word.Word, error) {
	args := m.Called(ctx, language, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*word.Word), args.Error(1)
}

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		AppName:  "Test App",
		HTTPPort: "8080",
	}

	logger := zerolog.New(zerolog.NewTestWriter(t))
	wordService := new(MockWordService)

	server := NewServer(cfg, logger, wordService)

	assert.NotNil(t, server)
	assert.Equal(t, cfg, server.cfg)
	assert.Equal(t, logger, server.log)
	assert.Equal(t, wordService, server.wordService)
	assert.NotNil(t, server.router)
}

func TestHealthCheck(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		AppName:  "Test App",
		HTTPPort: "8080",
	}

	// Create a test writer to capture log output
	buf := &bytes.Buffer{}
	logger := zerolog.New(buf)
	wordService := new(MockWordService)

	// Create server
	server := NewServer(cfg, logger, wordService)
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
	assert.Contains(t, buf.String(), "Health check request")
}

func TestLoggerMiddleware(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		AppName:  "Test App",
		HTTPPort: "8080",
	}

	// Create a test writer to capture log output
	buf := &bytes.Buffer{}
	logger := zerolog.New(buf)
	wordService := new(MockWordService)

	// Create server
	server := NewServer(cfg, logger, wordService)

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
	logOutput := buf.String()
	assert.Contains(t, logOutput, "Request started")
	assert.Contains(t, logOutput, "Request completed")
	assert.Contains(t, logOutput, "GET")
	assert.Contains(t, logOutput, "/test")
}

func TestSearchWord(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		AppName:  "Test App",
		HTTPPort: "8080",
	}

	logger := zerolog.New(zerolog.NewTestWriter(t))
	wordService := new(MockWordService)

	// Create server
	server := NewServer(cfg, logger, wordService)

	// Mock word service response
	testWord := &word.Word{
		Text:        "test",
		Language:    "en",
		Definitions: []string{"a procedure intended to establish quality"},
	}
	wordService.On("Search", mock.Anything, "test", "en").Return(testWord, nil)

	// Create request body
	requestBody := WordSearchRequest{
		Text:     "test",
		Language: "en",
	}
	jsonBody, _ := json.Marshal(requestBody)

	// Create test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/words/search", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Create router and register handler
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("logger", logger)
		c.Next()
	})
	router.POST("/api/v1/words/search", server.SearchWord)

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response WordSearchResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response.Word)
	assert.Equal(t, "test", response.Word.Text)
	assert.Equal(t, "en", response.Word.Language)
	assert.Equal(t, "a procedure intended to establish quality", response.Word.Definitions[0])

	// Verify mock was called
	wordService.AssertExpectations(t)
}

func TestGetRecentWords(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		AppName:  "Test App",
		HTTPPort: "8080",
	}

	logger := zerolog.New(zerolog.NewTestWriter(t))
	wordService := new(MockWordService)

	// Create server
	server := NewServer(cfg, logger, wordService)

	// Mock word service response
	testWords := []*word.Word{
		{Text: "test1", Language: "en"},
		{Text: "test2", Language: "en"},
	}
	wordService.On("GetRecentWords", mock.Anything, "en", 10).Return(testWords, nil)

	// Create test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/words/recent?language=en", nil)

	// Create router and register handler
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("logger", logger)
		c.Next()
	})
	router.GET("/api/v1/words/recent", server.GetRecentWords)

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response RecentWordsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response.Words)
	assert.Len(t, response.Words, 2)
	assert.Equal(t, "test1", response.Words[0].Text)
	assert.Equal(t, "test2", response.Words[1].Text)

	// Verify mock was called
	wordService.AssertExpectations(t)
}
