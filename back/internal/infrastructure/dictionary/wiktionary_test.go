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
	assert.NotNil(t, api.client)
	assert.Equal(t, "https://en.wiktionary.org/w/api.php", api.baseURL)
}

func TestFetchWord_Success(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request
		assert.Contains(t, r.URL.String(), "action=query")
		assert.Contains(t, r.URL.String(), "format=json")
		assert.Contains(t, r.URL.String(), "titles=test")

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"query": {
				"pages": {
					"12345": {
						"extract": "<p>Test is a procedure intended to establish the quality, performance, or reliability of something.</p>",
						"pronunciation": "tɛst",
						"translations": {
							"fr": "test",
							"es": "prueba"
						}
					}
				}
			}
		}`))
	}))
	defer server.Close()

	// Create API with mock server URL
	logger := zerolog.New(zerolog.NewTestWriter(t))
	api := NewWiktionaryAPI(logger)
	api.baseURL = server.URL

	// Execute
	ctx := context.Background()
	result, err := api.FetchWord(ctx, "test", "en")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test", result.Text)
	assert.Equal(t, "en", result.Language)
	assert.Contains(t, result.Definitions[0], "procedure intended to establish")
	assert.Equal(t, "tɛst", result.Pronunciation)
	assert.Equal(t, "test", result.Translations["fr"])
	assert.Equal(t, "prueba", result.Translations["es"])
}

func TestFetchWord_RequestError(t *testing.T) {
	// Setup with invalid URL to cause request error
	logger := zerolog.New(zerolog.NewTestWriter(t))
	api := NewWiktionaryAPI(logger)
	api.baseURL = "http://invalid-url-that-will-fail"

	// Execute
	ctx := context.Background()
	result, err := api.FetchWord(ctx, "test", "en")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to send request")
}

func TestFetchWord_InvalidResponse(t *testing.T) {
	// Setup mock server with invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	// Create API with mock server URL
	logger := zerolog.New(zerolog.NewTestWriter(t))
	api := NewWiktionaryAPI(logger)
	api.baseURL = server.URL

	// Execute
	ctx := context.Background()
	result, err := api.FetchWord(ctx, "test", "en")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to parse response")
}

func TestFetchWord_EmptyResponse(t *testing.T) {
	// Setup mock server with empty response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"query": {"pages": {}}}`))
	}))
	defer server.Close()

	// Create API with mock server URL
	logger := zerolog.New(zerolog.NewTestWriter(t))
	api := NewWiktionaryAPI(logger)
	api.baseURL = server.URL

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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
	}))
	defer server.Close()

	// Create API with mock server URL
	logger := zerolog.New(zerolog.NewTestWriter(t))
	api := NewWiktionaryAPI(logger)
	api.baseURL = server.URL

	// Execute
	ctx := context.Background()
	result, err := api.FetchWord(ctx, "test", "en")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "received non-200 response")
}

func TestExtractDefinitions(t *testing.T) {
	// Setup
	html := `<p>Definition 1.</p><p>Definition 2.</p><ul><li>Example 1</li><li>Example 2</li></ul>`

	// Execute
	definitions := extractDefinitions(html)

	// Assert
	assert.Len(t, definitions, 2)
	assert.Equal(t, "Definition 1.", definitions[0])
	assert.Equal(t, "Definition 2.", definitions[1])
}

func TestExtractExamples(t *testing.T) {
	// Setup
	html := `<p>Definition.</p><ul><li>Example 1</li><li>Example 2</li></ul>`

	// Execute
	examples := extractExamples(html)

	// Assert
	assert.Len(t, examples, 2)
	assert.Equal(t, "Example 1", examples[0])
	assert.Equal(t, "Example 2", examples[1])
}
