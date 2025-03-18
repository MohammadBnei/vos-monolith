package dictionary

import (
	"context"
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	wordDomain "voconsteroid/internal/domain/word"
	"voconsteroid/internal/infrastructure/dictionary/acl"
)

// MockWiktionaryScraper is a mock implementation of the WiktionaryScraper interface
type MockWiktionaryScraper struct {
	mock.Mock
}

func (m *MockWiktionaryScraper) FetchWordData(ctx context.Context, text, language string) (*acl.WiktionaryResponse, error) {
	args := m.Called(ctx, text, language)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*acl.WiktionaryResponse), args.Error(1)
}

func (m *MockWiktionaryScraper) FetchRelatedWordsData(ctx context.Context, word, language string) (*acl.WiktionaryRelatedResponse, error) {
	args := m.Called(ctx, word, language)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*acl.WiktionaryRelatedResponse), args.Error(1)
}

func (m *MockWiktionaryScraper) FetchSuggestionsData(ctx context.Context, prefix, language string) ([]string, error) {
	args := m.Called(ctx, prefix, language)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func TestNewWiktionaryAPI(t *testing.T) {
	// Setup
	logger := zerolog.New(zerolog.NewTestWriter(t))

	// Execute
	api := NewWiktionaryAPI(logger)

	// Assert
	assert.NotNil(t, api)
	assert.NotNil(t, api.scrapers)
	assert.NotNil(t, api.lookupServices["fr"]) // French lookup service should be registered
}

func TestWiktionaryAPI_FetchWord_SupportedLanguage(t *testing.T) {
	// Setup
	logger := zerolog.New(zerolog.NewTestWriter(t))
	api := NewWiktionaryAPI(logger)

	// Replace the scraper with a mock
	mockScraper := new(MockWiktionaryScraper)
	api.lookupServices["fr"].adapter = &acl.WiktionaryAdapter{logger: logger} // need to set adapter
	api.lookupServices["fr"].scraper = mockScraper

	// Setup expectations
	expectedResponse := &acl.WiktionaryResponse{
		Word:     "bonjour",
		Language: "fr",
	}
	mockScraper.On("FetchWordData", mock.Anything, "bonjour", "fr").Return(expectedResponse, nil)

	// Execute
	ctx := context.Background()
	result, err := api.FetchWord(ctx, "bonjour", "fr")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "bonjour", result.Text)
	assert.Equal(t, "fr", result.Language)
	mockScraper.AssertExpectations(t)
}

func TestWiktionaryAPI_FetchWord_UnsupportedLanguage(t *testing.T) {
	// Setup
	logger := zerolog.New(zerolog.NewTestWriter(t))
	api := NewWiktionaryAPI(logger)

	// Execute
	ctx := context.Background()
	result, err := api.FetchWord(ctx, "hello", "en")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, wordDomain.ErrWordNotFound))
	assert.Contains(t, err.Error(), "unsupported language en")
}

func TestWiktionaryAPI_FetchRelatedWords_SupportedLanguage(t *testing.T) {
	// Setup
	logger := zerolog.New(zerolog.NewTestWriter(t))
	api := NewWiktionaryAPI(logger)

	// Replace the scraper with a mock
	mockScraper := new(MockWiktionaryScraper)
	api.lookupServices["fr"].adapter = &acl.WiktionaryAdapter{logger: logger} // need to set adapter
	api.lookupServices["fr"].scraper = mockScraper

	// Setup expectations
	sourceWord := "bonjour"
	language := "fr"
	expectedResponse := &acl.WiktionaryRelatedResponse{
		SourceWord: sourceWord,
		Language:   language,
		Synonyms:   []string{"salut"},
		Antonyms:   []string{"au revoir"},
	}
	mockScraper.On("FetchRelatedWordsData", mock.Anything, sourceWord, language).Return(expectedResponse, nil)

	// Execute
	word := wordDomain.NewWord(sourceWord, language)
	ctx := context.Background()
	result, err := api.FetchRelatedWords(ctx, word)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, sourceWord, result.SourceWord.Text)
	assert.Equal(t, language, result.SourceWord.Language)
	assert.Equal(t, []string{"salut"}, convertWordSliceToStringSlice(result.Synonyms))
	assert.Equal(t, []string{"au revoir"}, convertWordSliceToStringSlice(result.Antonyms))
	mockScraper.AssertExpectations(t)
}

func TestWiktionaryAPI_FetchRelatedWords_UnsupportedLanguage(t *testing.T) {
	// Setup
	logger := zerolog.New(zerolog.NewTestWriter(t))
	api := NewWiktionaryAPI(logger)

	// Create a word with unsupported language
	sourceWord := wordDomain.NewWord("hello", "en")

	// Execute
	ctx := context.Background()
	result, err := api.FetchRelatedWords(ctx, sourceWord)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, wordDomain.ErrWordNotFound))
	assert.Contains(t, err.Error(), "unsupported language en")
}

func TestWiktionaryAPI_FetchSuggestions_SupportedLanguage(t *testing.T) {
	// Setup
	logger := zerolog.New(zerolog.NewTestWriter(t))
	api := NewWiktionaryAPI(logger)

	// Replace the scraper with a mock
	mockScraper := new(MockWiktionaryScraper)
	api.lookupServices["fr"].adapter = &acl.WiktionaryAdapter{logger: logger} // need to set adapter
	api.lookupServices["fr"].scraper = mockScraper

	// Setup expectations
	expectedSuggestions := []string{"test1", "test2"}
	mockScraper.On("FetchSuggestionsData", mock.Anything, "test", "fr").Return(expectedSuggestions, nil)

	// Execute
	ctx := context.Background()
	result, err := api.FetchSuggestions(ctx, "test", "fr")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedSuggestions, result)
	mockScraper.AssertExpectations(t)
}

func TestWiktionaryAPI_FetchSuggestions_UnsupportedLanguage(t *testing.T) {
	// Setup
	logger := zerolog.New(zerolog.NewTestWriter(t))
	api := NewWiktionaryAPI(logger)

	// Execute
	ctx := context.Background()
	result, err := api.FetchSuggestions(ctx, "test", "en")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, wordDomain.ErrWordNotFound))
	assert.Contains(t, err.Error(), "unsupported language en")
}

// Helper function to convert []*word.Word to []string for easier comparison
func convertWordSliceToStringSlice(words []*wordDomain.Word) []string {
	stringSlice := make([]string, len(words))
	for i, word := range words {
		stringSlice[i] = word.Text
	}
	return stringSlice
}
