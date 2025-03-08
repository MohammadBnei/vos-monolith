package dictionary

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewWiktionaryAPI(t *testing.T) {
	// Setup
	logger := zerolog.New(zerolog.NewTestWriter(t))

	// Execute
	api := NewWiktionaryAPI(logger)

	// Assert
	assert.NotNil(t, api)
}

func TestGetBaseURL(t *testing.T) {
	// Test cases
	testCases := []struct {
		language string
		expected string
	}{
		{"en", "https://en.wiktionary.org/wiki"},
		{"fr", "https://fr.wiktionary.org/wiki"},
		{"es", "https://es.wiktionary.org/wiki"},
		{"unknown", "https://en.wiktionary.org/wiki"}, // Default to English
	}

	// Execute and assert
	for _, tc := range testCases {
		result := defaultGetBaseURL(tc.language)
		assert.Equal(t, tc.expected, result, "For language %s", tc.language)
	}
}

func TestFetchWord_Success(t *testing.T) {
	// Create HTML with test data
	html := `
<!DOCTYPE html>
<html>
<head><title>Test - Wiktionary</title></head>
<body>
  <h1>Test</h1>
  <ol>
    <li>A procedure for critical evaluation; a means of determining the presence, quality, or truth of something.</li>
    <li>A chemical reaction used to identify or detect a particular substance.</li>
  </ol>
  <div class="example-needed">This word needs an example.</div>
  <ul class="citations">
    <li>The test was positive for COVID-19.</li>
  </ul>
  <span class="IPA">tɛst</span>
  <div class="translations">
    <li><span class="language" lang="fr">French</span>: <span class="translation">test</span></li>
    <li><span class="language" lang="es">Spanish</span>: <span class="translation">prueba</span></li>
  </div>
  <div id="Etymology">From Middle English test, from Old French test ("pot").</div>
</body>
</html>
`

	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}))
	defer server.Close()

	// Create a test-specific implementation to avoid actual web requests
	logger := zerolog.New(zerolog.NewTestWriter(t))
	api := &WiktionaryAPI{
		logger: logger.With().Str("component", "wiktionary_scraper").Logger(),
	}

	// Override the getBaseURL function for testing
	api.getBaseURL = func(language string) string {
		return server.URL
	}

	// Execute
	ctx := context.Background()
	result, err := api.FetchWord(ctx, "test", "en")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	if result == nil {
		return
	}
	assert.Equal(t, "test", result.Text)
	assert.Equal(t, "en", result.Language)

	// Check definitions
	assert.GreaterOrEqual(t, len(result.Definitions), 1)

	// Check examples
	assert.GreaterOrEqual(t, len(result.Examples), 1)

	// Check pronunciation
	assert.Equal(t, "tɛst", result.Pronunciation)

	// Check etymology
	assert.Contains(t, result.Etymology, "Middle English")
}

func TestFetchWord_RequestError(t *testing.T) {
	// Setup
	logger := zerolog.New(zerolog.NewTestWriter(t))
	api := NewWiktionaryAPI(logger)

	// Override the getBaseURL function for testing
	api.getBaseURL = func(language string) string {
		return "http://invalid-url-that-will-fail"
	}

	// Execute
	ctx := context.Background()
	result, err := api.FetchWord(ctx, "test", "en")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to visit page")
}

func TestFetchWord_EmptyResponse(t *testing.T) {
	// Create empty HTML
	html := `
<!DOCTYPE html>
<html>
<head><title>Test - Wiktionary</title></head>
<body>
  <h1>Test</h1>
  <!-- No definitions -->
</body>
</html>
`

	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}))
	defer server.Close()

	// Create a test-specific implementation
	logger := zerolog.New(zerolog.NewTestWriter(t))
	api := &WiktionaryAPI{
		logger: logger.With().Str("component", "wiktionary_scraper").Logger(),
	}

	// Override the getBaseURL function for testing
	api.getBaseURL = func(language string) string {
		return server.URL
	}

	// Execute
	ctx := context.Background()
	result, err := api.FetchWord(ctx, "test", "en")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no word data found")
}

func TestFetchWord_ErrorResponse(t *testing.T) {
	// Setup mock server with error response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	// Create a test-specific implementation
	logger := zerolog.New(zerolog.NewTestWriter(t))
	api := &WiktionaryAPI{
		logger: logger.With().Str("component", "wiktionary_scraper").Logger(),
	}

	// Override the getBaseURL method for testing
	originalGetBaseURL := api.getBaseURL
	defer func() { api.getBaseURL = originalGetBaseURL }()
	api.getBaseURL = func(language string) string {
		return server.URL
	}

	// Execute
	ctx := context.Background()
	result, err := api.FetchWord(ctx, "test", "en")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to visit page")
}
