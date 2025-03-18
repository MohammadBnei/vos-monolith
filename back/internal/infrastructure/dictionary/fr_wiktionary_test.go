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
func createTestAPI(t *testing.T) *FrenchWiktionaryScraper {
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	return NewFrenchWiktionaryScraper(logger)
}

func TestFrenchWiktionaryScraper_FetchWord_Success(t *testing.T) {
	api := createTestAPI(t)

	testCases := []struct {
		name     string
		word     string
		language string
	}{
		{"common noun", "maison", "fr"},
		{"verb", "manger", "fr"},
		{"adjective", "grand", "fr"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			word, err := api.FetchWordData(context.Background(), tc.word, tc.language)
			assert.NoError(t, err)
			require.NotNil(t, word)

			// Basic validation
			assert.Equal(t, tc.word, word.Word)
			assert.Equal(t, tc.language, word.Language)
			assert.NotEmpty(t, word.Definitions)

			// Validate definitions
			for _, def := range word.Definitions {
				assert.NotEmpty(t, def.Text)
				if def.WordType != "" {
					assert.True(t, french.IsValidWordType(french.WordType(def.WordType)))
				}
				if def.Gender != "" {
					assert.True(t, french.IsValidGender(french.Gender(def.Gender)))
				}
			}

			// Validate search terms
			assert.Contains(t, word.SearchTerms, tc.word)
		})
	}
}

func TestFrenchWiktionaryScraper_FetchWord_ErrorCases(t *testing.T) {
	api := createTestAPI(t)

	testCases := []struct {
		name        string
		word        string
		language    string
		expectedErr error
	}{
		{"empty word", "", "fr", wordDomain.ErrWordNotFound},
		{"unsupported language", "maison", "en", wordDomain.ErrWordNotFound},
		{"non-existent word", "nonexistentword12345", "fr", wordDomain.ErrWordNotFound},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			word, err := api.FetchWord(context.Background(), tc.word, tc.language)
			assert.Error(t, err)
			assert.Nil(t, word)
			assert.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

func TestFrenchWiktionaryScraper_FetchWord_ContextCancellation(t *testing.T) {
	api := createTestAPI(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	word, err := api.FetchWord(ctx, "maison", "fr")
	assert.Error(t, err)
	assert.Nil(t, word)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestFrenchWiktionaryScraper_FetchWord_EmptyHTML(t *testing.T) {
	// Create a test server that returns empty HTML
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body></body></html>"))
	}))
	defer server.Close()

	// Create a logger
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	// Create a French Wiktionary API with a custom getBaseURL function
	api := NewFrenchWiktionaryScraper(logger)
	api.getBaseURL = func() string {
		return server.URL
	}

	// Fetch the word - should return error because no definitions found
	_, err := api.FetchWordData(context.Background(), "test", "fr")
	assert.Error(t, err, "Should return error for empty HTML")
	assert.Contains(t, err.Error(), "no French section found", "Error should be about word not found")
	assert.ErrorIs(t, err, wordDomain.ErrWordNotFound, "Error should be ErrWordNotFound")
}

// func TestFrenchWiktionaryScraper_FetchRelatedWords(t *testing.T) {
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

func TestFrenchWiktionaryScraper_FetchRelatedWords_ContextCancellation(t *testing.T) {
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

func TestFrenchWiktionaryScraper_FetchSuggestions(t *testing.T) {
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

func TestFrenchWiktionaryScraper_FetchSuggestions_UnsupportedLanguage(t *testing.T) {
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

func TestFrenchWiktionaryScraper_FetchSuggestions_MockServer(t *testing.T) {
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
	api := &FrenchWiktionaryScraper{
		logger: logger.With().Str("component", "fr_wiktionary_scraper").Logger(),
		getBaseURL: func() string {
			return server.URL
		},
	}

	// Test the FetchSuggestions method
	ctx := context.Background()
	suggestions, err := api.FetchSuggestionsData(ctx, "test", "fr")

	// Assert
	assert.NoError(t, err)
	assert.Len(t, suggestions, 2)
	assert.Equal(t, "test1", suggestions[0])
	assert.Equal(t, "test2", suggestions[1])
}

func TestFrenchWiktionaryScraper_FetchSuggestions_InvalidJSON(t *testing.T) {
	// Create a test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	// Create a logger
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	// Create a French Wiktionary API with a custom HTTP client
	api := &FrenchWiktionaryScraper{
		logger: logger.With().Str("component", "fr_wiktionary_scraper").Logger(),
		getBaseURL: func() string {
			return server.URL
		},
	}

	// Test the FetchSuggestions method
	ctx := context.Background()
	suggestions, err := api.FetchSuggestionsData(ctx, "test", "fr")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, suggestions)
	assert.Contains(t, err.Error(), "failed to decode response")
}

func TestFrenchWiktionaryScraper_FetchSuggestions_ServerError(t *testing.T) {
	// Create a test server that returns a server error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create a logger
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	// Create a French Wiktionary API with a custom HTTP client
	api := &FrenchWiktionaryScraper{
		logger: logger.With().Str("component", "fr_wiktionary_scraper").Logger(),
		getBaseURL: func() string {
			return server.URL
		},
	}

	// Test the FetchSuggestions method
	ctx := context.Background()
	suggestions, err := api.FetchSuggestionsData(ctx, "test", "fr")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, suggestions)
	assert.Contains(t, err.Error(), "received non-OK status code")
}
func TestFrenchWiktionaryScraper_FetchWord_EdgeCases(t *testing.T) {
	api := createTestAPI(t)

	testCases := []struct {
		name     string
		word     string
		language string
	}{
		{"word with diacritics", "éclair", "fr"},
		{"compound word", "porte-monnaie", "fr"},
		{"word with apostrophe", "aujourd'hui", "fr"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			word, err := api.FetchWordData(context.Background(), tc.word, tc.language)
			assert.NoError(t, err)
			require.NotNil(t, word)

			assert.Equal(t, tc.word, word.Word)
			assert.Equal(t, tc.language, word.Language)
			assert.NotEmpty(t, word.Definitions)
		})
	}
}

func TestFrenchWiktionaryScraper_RealFetchWord_NotFound(t *testing.T) {
	// Create a test API with real Wiktionary URL
	api := createTestAPI(t)

	// Test with a non-existent word
	_, err := api.FetchWordData(context.Background(), "nonexistentword12345", "fr")
	assert.Error(t, err)
	assert.ErrorIs(t, err, wordDomain.ErrWordNotFound)
}
func TestFrenchWiktionaryScraper_DetermineWordType(t *testing.T) {
	api := &FrenchWiktionaryScraper{}

	testCases := []struct {
		sectionTitle string
		expectedType string
	}{
		{"Nom commun", string(french.Noun)},
		{"Verbe", string(french.Verb)},
		{"Adjectif", string(french.Adjective)},
		{"Adverbe", string(french.Adverb)},
		{"Pronom", string(french.Pronoun)},
		{"Préposition", string(french.Preposition)},
		{"Conjonction", string(french.Conjunction)},
		{"Interjection", string(french.Interjection)},
		{"Unknown", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.sectionTitle, func(t *testing.T) {
			result := api.determineWordType(tc.sectionTitle)
			assert.Equal(t, tc.expectedType, result)
		})
	}
}
func TestFrenchWiktionaryScraper_MapLanguageNameToCode(t *testing.T) {
	api := &FrenchWiktionaryScraper{}

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
	def.SetWordType("nom")
	def.AddExample("This is a test example")
	def.SetGender("masculine")
	def.SetPronunciation("/tɛst/")
	def.AddLanguageSpecific("plural", "tests")
	def.AddNote("This is a test note")

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
	assert.Equal(t, "nom", word.GetPrimaryWordType())

	// Test GetAllSpecifics
	specifics := word.GetAllSpecifics()
	assert.Contains(t, specifics, "tests")

	// Test GetDefinitionsByType
	nounDefs := word.GetDefinitionsByType("nom")
	assert.Len(t, nounDefs, 1)
	assert.Equal(t, "A test definition", nounDefs[0].Text)

	verbDefs := word.GetDefinitionsByType("verb")
	assert.Len(t, verbDefs, 0)

	// Test ValidateDefinition with valid values
	validDef := wordDomain.NewDefinition()
	validDef.SetWordType("nom")
	validDef.SetGender("féminin") // Use a valid gender for French
	err := word.ValidateDefinition(validDef)
	assert.NoError(t, err)

	// Test with invalid word type
	invalidTypeDef := wordDomain.NewDefinition()
	invalidTypeDef.SetWordType("invalid_type")
	err = word.ValidateDefinition(invalidTypeDef)
	assert.Error(t, err)
	assert.ErrorIs(t, err, wordDomain.ErrInvalidWordType)

	// Test with invalid gender
	invalidGenderDef := wordDomain.NewDefinition()
	invalidGenderDef.SetWordType("nom") // Set a valid word type
	invalidGenderDef.SetGender("invalid_gender")
	err = word.ValidateDefinition(invalidGenderDef)
	assert.Error(t, err)
	assert.ErrorIs(t, err, wordDomain.ErrInvalidGender)
}
