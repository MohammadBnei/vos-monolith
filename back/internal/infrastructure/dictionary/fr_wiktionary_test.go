package dictionary

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wordDomain "voconsteroid/internal/domain/word"
)

// Helper function to create a test API with real Wiktionary URL
func createTestAPI(t *testing.T) *FrenchWiktionaryAPI {
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	return NewFrenchWiktionaryAPI(logger)
}

func TestFrenchWiktionaryAPI_FetchWord(t *testing.T) {
	// Create a test API with real Wiktionary URL
	api := createTestAPI(t)

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

	// Check that the timestamps are set
	assert.False(t, word.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.False(t, word.UpdatedAt.IsZero(), "UpdatedAt should be set")
}

func TestFrenchWiktionaryAPI_FetchWord_NotFound(t *testing.T) {
	// Create a test API with real Wiktionary URL
	api := createTestAPI(t)

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
	assert.Contains(t, err.Error(), "word not found", "Error should be about word not found")
	assert.ErrorIs(t, err, wordDomain.ErrWordNotFound, "Error should be ErrWordNotFound")
}

// func TestFrenchWiktionaryAPI_FetchRelatedWords(t *testing.T) {
// 	// Create a test API with real Wiktionary URL
// 	api := createTestAPI(t)

// 	// Create a test word with synonyms and antonyms
// 	testWord := wordDomain.NewWord("test", "fr")
// 	testWord.AddSynonym("synonym1")
// 	testWord.AddSynonym("synonym2")
// 	testWord.AddAntonym("antonym1")

// 	// Create a context
// 	ctx := context.Background()

// 	// Test the FetchRelatedWords method
// 	relatedWords, err := api.FetchRelatedWords(ctx, testWord)

// 	// We expect an error because the mock server doesn't return valid HTML
// 	// but we should still get a RelatedWords object with minimal words
// 	assert.Error(t, err)
// 	require.NotNil(t, relatedWords)
// 	assert.Equal(t, testWord, relatedWords.SourceWord)
// 	assert.Len(t, relatedWords.Synonyms, 2)
// 	assert.Len(t, relatedWords.Antonyms, 1)

// 	// Check that the minimal words were created correctly
// 	assert.Equal(t, "synonym1", relatedWords.Synonyms[0].Text)
// 	assert.Equal(t, "synonym2", relatedWords.Synonyms[1].Text)
// 	assert.Equal(t, "antonym1", relatedWords.Antonyms[0].Text)
// }

func TestFrenchWiktionaryAPI_FetchRelatedWords_ContextCancellation(t *testing.T) {
	// Create a test API with real Wiktionary URL
	api := createTestAPI(t)

	// Create a test word
	testWord := wordDomain.NewWord("test", "fr")
	testWord.AddSynonym("synonym1")
	testWord.AddAntonym("antonym1")

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Test the FetchRelatedWords method with a cancelled context
	_, err := api.FetchRelatedWords(ctx, testWord)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}
func TestFrenchWiktionaryAPI_RealFetchWord(t *testing.T) {
	// Create a test API with real Wiktionary URL
	api := createTestAPI(t)

	// Test with a real French word
	word, err := api.FetchWord(context.Background(), "maison", "fr")
	assert.NoError(t, err)
	require.NotNil(t, word, "Word should not be nil")

	// Verify basic word data
	assert.Equal(t, "maison", word.Text)
	assert.Equal(t, "fr", word.Language)

	// Check that we have definitions
	assert.Greater(t, len(word.Definitions), 0, "Should have at least one definition")

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

	// Check that the timestamps are set
	assert.False(t, word.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.False(t, word.UpdatedAt.IsZero(), "UpdatedAt should be set")
}

func TestFrenchWiktionaryAPI_RealFetchWord_NotFound(t *testing.T) {
	// Create a test API with real Wiktionary URL
	api := createTestAPI(t)

	// Test with a non-existent word
	_, err := api.FetchWord(context.Background(), "nonexistentword12345", "fr")
	assert.Error(t, err)
	assert.ErrorIs(t, err, wordDomain.ErrWordNotFound)
}
func TestFrenchWiktionaryAPI_DetermineWordType(t *testing.T) {
	api := &FrenchWiktionaryAPI{}

	testCases := []struct {
		sectionTitle string
		expectedType string
	}{
		{"Nom commun", "noun"},
		{"Verbe", "verb"},
		{"Adjectif", "adjective"},
		{"Adverbe", "adverb"},
		{"Pronom", "pronoun"},
		{"Préposition", "preposition"},
		{"Conjonction", "conjunction"},
		{"Interjection", "interjection"},
		{"Unknown", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.sectionTitle, func(t *testing.T) {
			result := api.determineWordType(tc.sectionTitle)
			assert.Equal(t, tc.expectedType, result)
		})
	}
}
func TestFrenchWiktionaryAPI_MapLanguageNameToCode(t *testing.T) {
	api := &FrenchWiktionaryAPI{}

	testCases := []struct {
		langName string
		expected string
	}{
		{"Allemand", "de"},
		{"Anglais", "en"},
		{"Espagnol", "es"},
		{"Italien", "it"},
		{"Portugais", "pt"},
		{"Roumain", "ro"},
		{"Unknown", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.langName, func(t *testing.T) {
			result := api.mapLanguageNameToCode(tc.langName)
			assert.Equal(t, tc.expected, result)
		})
	}
}
func TestContainsString(t *testing.T) {
	testCases := []struct {
		name     string
		slice    []string
		str      string
		expected bool
	}{
		{"empty slice", []string{}, "test", false},
		{"string not in slice", []string{"a", "b", "c"}, "test", false},
		{"string in slice", []string{"a", "test", "c"}, "test", true},
		{"string at beginning", []string{"test", "b", "c"}, "test", true},
		{"string at end", []string{"a", "b", "test"}, "test", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := containsString(tc.slice, tc.str)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
