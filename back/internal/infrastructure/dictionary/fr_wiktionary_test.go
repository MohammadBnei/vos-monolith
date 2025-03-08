package dictionary

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"voconsteroid/internal/domain/word"
)

func TestFrenchWiktionaryAPI_FetchWord(t *testing.T) {
	// Load test HTML
	html, err := os.ReadFile("fr_wiktionary_test.html")
	if err != nil {
		t.Fatalf("Failed to read test HTML file: %v", err)
	}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(html)
	}))
	defer server.Close()

	// Create a logger
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	// Create a French Wiktionary API with a custom getBaseURL function
	api := NewFrenchWiktionaryAPI(logger)
	api.getBaseURL = func(language string) string {
		return server.URL
	}

	// Fetch the word
	word, err := api.FetchWord(context.Background(), "test", "fr")
	assert.NoError(t, err)
	require.NotNil(t, word, "Word should not be nil")

	// Verify the word data
	assert.Equal(t, "test", word.Text)
	assert.Equal(t, "fr", word.Language)

	// Check that we have definitions
	assert.Greater(t, len(word.Definitions), 0, "Should have at least one definition")
	
	// Check the first definition
	if len(word.Definitions) > 0 {
		def := word.Definitions[0]
		assert.NotEmpty(t, def.Text, "Definition text should not be empty")
		// If the definition has examples, check them
		if len(def.Examples) > 0 {
			assert.NotEmpty(t, def.Examples[0], "Example should not be empty")
		}
	}

	// Check that we have examples
	assert.Greater(t, len(word.Examples), 0, "Should have at least one example")

	// Check that we have synonyms
	assert.Greater(t, len(word.Synonyms), 0, "Should have at least one synonym")

	// Check that we have a pronunciation
	assert.NotEmpty(t, word.Pronunciation, "Should have pronunciation")
	if word.Pronunciation != nil {
		assert.NotEmpty(t, word.Pronunciation["ipa"], "Should have IPA pronunciation")
	}

	// Check that we have an etymology
	assert.NotEmpty(t, word.Etymology, "Should have etymology")

	// Check that we have a plural form
	assert.Contains(t, word.Translations, "plural", "Should have plural form in translations")
	
	// Check that we have word forms
	assert.Greater(t, len(word.Forms), 0, "Should have at least one word form")
	
	// Check that the timestamps are set
	assert.False(t, word.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.False(t, word.UpdatedAt.IsZero(), "UpdatedAt should be set")
}

func TestFrenchWiktionaryAPI_FetchWord_Timeout(t *testing.T) {
	// Create a test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) // Delay response
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body>Test page</body></html>"))
	}))
	defer server.Close()

	// Create a logger
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	// Create a French Wiktionary API with a custom getBaseURL function
	api := NewFrenchWiktionaryAPI(logger)
	api.getBaseURL = func(language string) string {
		return server.URL
	}

	// Create a context with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Fetch the word - should timeout
	_, err := api.FetchWord(ctx, "test", "fr")
	assert.Error(t, err, "Should return error due to timeout")
	assert.Contains(t, err.Error(), "context deadline exceeded", "Error should be about context deadline")
}

func TestFrenchWiktionaryAPI_FetchWord_NotFound(t *testing.T) {
	// Create a test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("<html><body>Page not found</body></html>"))
	}))
	defer server.Close()

	// Create a logger
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	// Create a French Wiktionary API with a custom getBaseURL function
	api := NewFrenchWiktionaryAPI(logger)
	api.getBaseURL = func(language string) string {
		return server.URL
	}

	// Fetch the word - should return not found error
	_, err := api.FetchWord(context.Background(), "nonexistentword", "fr")
	assert.Error(t, err, "Should return error for non-existent word")
}

func TestFrenchWiktionaryAPI_FetchWord_EmptyHTML(t *testing.T) {
	// Create a test server that returns empty HTML
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body></body></html>"))
	}))
	defer server.Close()

	// Create a logger
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	// Create a French Wiktionary API with a custom getBaseURL function
	api := NewFrenchWiktionaryAPI(logger)
	api.getBaseURL = func(language string) string {
		return server.URL
	}

	// Fetch the word - should return error because no definitions found
	_, err := api.FetchWord(context.Background(), "test", "fr")
	assert.Error(t, err, "Should return error for empty HTML")
	assert.Contains(t, err.Error(), "no word data found", "Error should be about no word data found")
	assert.ErrorIs(t, err, word.ErrWordNotFound, "Error should be ErrWordNotFound")
}
