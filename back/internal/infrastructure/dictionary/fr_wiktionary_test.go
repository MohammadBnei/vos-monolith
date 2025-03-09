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
	"voconsteroid/internal/domain/word/languages/french"
)

// Helper function to create a test API with real Wiktionary URL
func createTestAPI(t *testing.T) *FrenchWiktionaryAPI {
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	return NewFrenchWiktionaryAPI(logger)
}

func TestFrenchWiktionaryAPI_FetchWord(t *testing.T) {
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

	// Check the first definition
	if len(word.Definitions) > 0 {
		def := word.Definitions[0]
		assert.NotEmpty(t, def.Text, "Definition text should not be empty")

		// These fields might not always be populated depending on the word
		// So we don't assert they're not empty, just check them if they exist
		if def.Pronunciation != "" {
			t.Logf("Found pronunciation: %s", def.Pronunciation)
		}

		if len(def.LanguageSpecifics) > 0 {
			t.Logf("Found language specifics: %v", def.LanguageSpecifics)
		}

		// If the definition has examples, check them
		if len(def.Examples) > 0 {
			assert.NotEmpty(t, def.Examples[0], "Example should not be empty")
		}

		// Validate French-specific fields
		if def.WordType != "" {
			assert.True(t, french.IsValidWordType(french.WordType(def.WordType)),
				"Word type should be valid for French")
		}
		if def.Gender != "" {
			assert.True(t, french.IsValidGender(french.Gender(def.Gender)),
				"Gender should be valid for French")
		}

		// Check for notes in definition
		if len(def.Notes) > 0 {
			assert.NotEmpty(t, def.Notes[0], "Note should not be empty")
		}
	}

	// Check that we have synonyms
	assert.Greater(t, len(word.Synonyms), 0, "Should have at least one synonym")

	// Check that we have an etymology
	assert.NotEmpty(t, word.Etymology, "Should have etymology")

	// Check search terms
	assert.NotEmpty(t, word.SearchTerms, "SearchTerms should be populated")
	assert.Contains(t, word.SearchTerms, "maison", "Main word should be in search terms")

	// Lemma might not always be set
	if word.Lemma != "" {
		t.Logf("Found lemma: %s", word.Lemma)
	}

	// Check translations if available
	if len(word.Translations) > 0 {
		for lang, translation := range word.Translations {
			assert.NotEmpty(t, lang, "Translation language should not be empty")
			assert.NotEmpty(t, translation, "Translation text should not be empty")
		}
	}

	// Check antonyms if available
	if len(word.Antonyms) > 0 {
		assert.NotEmpty(t, word.Antonyms[0], "Antonym should not be empty")
	}

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

func TestFrenchWiktionaryAPI_FetchSuggestions(t *testing.T) {
	// Skip this test in CI environments or when running automated tests
	if testing.Short() {
		t.Skip("Skipping test that makes external API calls")
	}

	// Create a test API
	api := createTestAPI(t)

	// Test with a real French prefix
	ctx := context.Background()
	suggestions, err := api.FetchSuggestions(ctx, "mai", "fr")

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, suggestions)
	
	// Check that we got string suggestions
	for _, suggestion := range suggestions {
		assert.NotEmpty(t, suggestion)
	}
}

func TestFrenchWiktionaryAPI_FetchSuggestions_UnsupportedLanguage(t *testing.T) {
	// Create a test API
	api := createTestAPI(t)

	// Test with an unsupported language
	ctx := context.Background()
	suggestions, err := api.FetchSuggestions(ctx, "test", "en")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, suggestions)
	assert.Contains(t, err.Error(), "unsupported language en")
}

func TestFrenchWiktionaryAPI_FetchSuggestions_MockServer(t *testing.T) {
	// Create a test server that returns a mock JSON response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"pages": [
				{"id": 1, "key": "test1", "title": "test1", "excerpt": "test1"},
				{"id": 2, "key": "test2", "title": "test2", "excerpt": "test2"}
			]
		}`))
	}))
	defer server.Close()

	// Create a logger
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	// Create a French Wiktionary API with a custom HTTP client
	api := &FrenchWiktionaryAPI{
		logger: logger.With().Str("component", "fr_wiktionary_scraper").Logger(),
		getBaseURL: func(language string) string {
			return server.URL
		},
	}

	// Test the FetchSuggestions method
	ctx := context.Background()
	suggestions, err := api.FetchSuggestions(ctx, "test", "fr")

	// Assert
	assert.NoError(t, err)
	assert.Len(t, suggestions, 2)
	assert.Equal(t, "test1", suggestions[0])
	assert.Equal(t, "test2", suggestions[1])
}

func TestFrenchWiktionaryAPI_FetchSuggestions_InvalidJSON(t *testing.T) {
	// Create a test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	// Create a logger
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	// Create a French Wiktionary API with a custom HTTP client
	api := &FrenchWiktionaryAPI{
		logger: logger.With().Str("component", "fr_wiktionary_scraper").Logger(),
		getBaseURL: func(language string) string {
			return server.URL
		},
	}

	// Test the FetchSuggestions method
	ctx := context.Background()
	suggestions, err := api.FetchSuggestions(ctx, "test", "fr")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, suggestions)
	assert.Contains(t, err.Error(), "failed to decode response")
}

func TestFrenchWiktionaryAPI_FetchSuggestions_ServerError(t *testing.T) {
	// Create a test server that returns a server error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create a logger
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	// Create a French Wiktionary API with a custom HTTP client
	api := &FrenchWiktionaryAPI{
		logger: logger.With().Str("component", "fr_wiktionary_scraper").Logger(),
		getBaseURL: func(language string) string {
			return server.URL
		},
	}

	// Test the FetchSuggestions method
	ctx := context.Background()
	suggestions, err := api.FetchSuggestions(ctx, "test", "fr")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, suggestions)
	assert.Contains(t, err.Error(), "received non-OK status code")
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

	// Check that we have synonyms
	assert.Greater(t, len(word.Synonyms), 0, "Should have at least one synonym")

	// Check that we have an etymology
	assert.NotEmpty(t, word.Etymology, "Should have etymology")

	// Check search terms
	assert.NotEmpty(t, word.SearchTerms, "SearchTerms should be populated")
	assert.Contains(t, word.SearchTerms, "maison", "Main word should be in search terms")

	// Check lemma if available
	if word.Lemma != "" {
		assert.NotEmpty(t, word.Lemma, "Lemma should not be empty if set")
	}

	// Check translations if available
	if len(word.Translations) > 0 {
		for lang, translation := range word.Translations {
			assert.NotEmpty(t, lang, "Translation language should not be empty")
			assert.NotEmpty(t, translation, "Translation text should not be empty")
		}
	}

	// Check antonyms if available
	if len(word.Antonyms) > 0 {
		assert.NotEmpty(t, word.Antonyms[0], "Antonym should not be empty")
	}

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

// TestWord_EntityMethods tests the methods of the Word entity
func TestWord_EntityMethods(t *testing.T) {
	// Create a new word
	word := wordDomain.NewWord("test", "fr")

	// Test SetLemma
	word.SetLemma("tester")
	assert.Equal(t, "tester", word.Lemma)
	assert.Contains(t, word.SearchTerms, "tester")

	// Test AddDefinition
	def := wordDomain.NewDefinition()
	def.Text = "A test definition"
	def.WordType = "noun"
	def.Examples = []string{"This is a test example"}
	def.Gender = "masculine"
	def.Pronunciation = "/tɛst/"
	def.LanguageSpecifics = map[string]string{"plural": "tests"}
	def.Notes = []string{"This is a test note"}

	word.AddDefinition(def)
	assert.Len(t, word.Definitions, 1)
	assert.Equal(t, "A test definition", word.Definitions[0].Text)
	assert.Contains(t, word.SearchTerms, "tests")

	// Test AddSearchTerm
	word.AddSearchTerm("testing")
	assert.Contains(t, word.SearchTerms, "testing")

	// Test adding duplicate search term
	word.AddSearchTerm("testing")
	count := 0
	for _, term := range word.SearchTerms {
		if term == "testing" {
			count++
		}
	}
	assert.Equal(t, 1, count, "Duplicate search term should not be added")

	// Test AddSynonym
	word.AddSynonym("examination")
	assert.Contains(t, word.Synonyms, "examination")

	// Test adding duplicate synonym
	word.AddSynonym("examination")
	count = 0
	for _, syn := range word.Synonyms {
		if syn == "examination" {
			count++
		}
	}
	assert.Equal(t, 1, count, "Duplicate synonym should not be added")

	// Test AddAntonym
	word.AddAntonym("ignorance")
	assert.Contains(t, word.Antonyms, "ignorance")

	// Test adding duplicate antonym
	word.AddAntonym("ignorance")
	count = 0
	for _, ant := range word.Antonyms {
		if ant == "ignorance" {
			count++
		}
	}
	assert.Equal(t, 1, count, "Duplicate antonym should not be added")

	// Test GetPrimaryWordType
	assert.Equal(t, "noun", word.GetPrimaryWordType())

	// Test GetAllSpecifics
	specifics := word.GetAllSpecifics()
	assert.Contains(t, specifics, "tests")

	// Test GetDefinitionsByType
	nounDefs := word.GetDefinitionsByType("noun")
	assert.Len(t, nounDefs, 1)
	assert.Equal(t, "A test definition", nounDefs[0].Text)

	verbDefs := word.GetDefinitionsByType("verb")
	assert.Len(t, verbDefs, 0)

	// Test ValidateDefinition with valid values
	validDef := wordDomain.NewDefinition()
	validDef.WordType = "nom"
	validDef.Gender = "féminin" // Use a valid gender for French
	err := word.ValidateDefinition(validDef)
	assert.NoError(t, err)

	// Test with invalid word type
	invalidTypeDef := wordDomain.NewDefinition()
	invalidTypeDef.WordType = "invalid_type"
	err = word.ValidateDefinition(invalidTypeDef)
	assert.Error(t, err)
	assert.ErrorIs(t, err, wordDomain.ErrInvalidWordType)

	// Test with invalid gender
	invalidGenderDef := wordDomain.NewDefinition()
	invalidGenderDef.WordType = "nom" // Set a valid word type
	invalidGenderDef.Gender = "invalid_gender"
	err = word.ValidateDefinition(invalidGenderDef)
	assert.Error(t, err)
	assert.ErrorIs(t, err, wordDomain.ErrInvalidGender)
}
