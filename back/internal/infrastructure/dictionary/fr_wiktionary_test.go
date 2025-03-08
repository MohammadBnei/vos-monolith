package dictionary

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
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
	assert.NotNil(t, word)

	// Verify the word data
	assert.Equal(t, "test", word.Text)
	assert.Equal(t, "fr", word.Language)
	
	// Check that we have definitions
	assert.Greater(t, len(word.Definitions), 0)
	
	// Check that we have examples
	assert.Greater(t, len(word.Examples), 0)
	
	// Check that we have synonyms
	assert.Greater(t, len(word.Synonyms), 0)
	
	// Check that we have a pronunciation
	assert.NotEmpty(t, word.Pronunciation)
	
	// Check that we have an etymology
	assert.NotEmpty(t, word.Etymology)
	
	// Check that we have a plural form
	assert.Contains(t, word.Translations, "plural")
}
