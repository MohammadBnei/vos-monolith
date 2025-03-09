package dictionary

import (
	"context"
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	wordDomain "voconsteroid/internal/domain/word"
)

// MockDictionaryAPI is a mock implementation of the DictionaryAPI interface
type MockDictionaryAPI struct {
	mock.Mock
}

func (m *MockDictionaryAPI) FetchWord(ctx context.Context, text, language string) (*wordDomain.Word, error) {
	args := m.Called(ctx, text, language)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wordDomain.Word), args.Error(1)
}

func (m *MockDictionaryAPI) FetchRelatedWords(ctx context.Context, word *wordDomain.Word) (*wordDomain.RelatedWords, error) {
	args := m.Called(ctx, word)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wordDomain.RelatedWords), args.Error(1)
}

func (m *MockDictionaryAPI) FetchSuggestions(ctx context.Context, prefix, language string) ([]string, error) {
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
	assert.NotNil(t, api.scrapers["fr"]) // French scraper should be registered
}

func TestWiktionaryAPI_FetchWord_SupportedLanguage(t *testing.T) {
	// Setup
	logger := zerolog.New(zerolog.NewTestWriter(t))
	api := &WiktionaryAPI{
		logger:   logger,
		scrapers: make(map[string]wordDomain.DictionaryAPI),
	}

	// Create mock scraper
	mockScraper := new(MockDictionaryAPI)
	api.scrapers["fr"] = mockScraper

	// Setup expectations
	expectedWord := wordDomain.NewWord("bonjour", "fr")
	mockScraper.On("FetchWord", mock.Anything, "bonjour", "fr").Return(expectedWord, nil)

	// Execute
	ctx := context.Background()
	result, err := api.FetchWord(ctx, "bonjour", "fr")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedWord, result)
	mockScraper.AssertExpectations(t)
}

func TestWiktionaryAPI_FetchWord_UnsupportedLanguage(t *testing.T) {
	// Setup
	logger := zerolog.New(zerolog.NewTestWriter(t))
	api := &WiktionaryAPI{
		logger:   logger,
		scrapers: make(map[string]wordDomain.DictionaryAPI),
	}

	// Only register French
	mockScraper := new(MockDictionaryAPI)
	api.scrapers["fr"] = mockScraper

	// Execute
	ctx := context.Background()
	result, err := api.FetchWord(ctx, "hello", "en")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, wordDomain.ErrWordNotFound))
	assert.Contains(t, err.Error(), "unsupported language en")
	mockScraper.AssertNotCalled(t, "FetchWord")
}

func TestWiktionaryAPI_FetchRelatedWords_SupportedLanguage(t *testing.T) {
	// Setup
	logger := zerolog.New(zerolog.NewTestWriter(t))
	api := &WiktionaryAPI{
		logger:   logger,
		scrapers: make(map[string]wordDomain.DictionaryAPI),
	}

	// Create mock scraper
	mockScraper := new(MockDictionaryAPI)
	api.scrapers["fr"] = mockScraper

	// Setup expectations
	sourceWord := wordDomain.NewWord("bonjour", "fr")
	expectedRelated := &wordDomain.RelatedWords{
		SourceWord: sourceWord,
		Synonyms:   []*wordDomain.Word{wordDomain.NewWord("salut", "fr")},
		Antonyms:   []*wordDomain.Word{wordDomain.NewWord("au revoir", "fr")},
	}
	mockScraper.On("FetchRelatedWords", mock.Anything, sourceWord).Return(expectedRelated, nil)

	// Execute
	ctx := context.Background()
	result, err := api.FetchRelatedWords(ctx, sourceWord)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedRelated, result)
	mockScraper.AssertExpectations(t)
}

func TestWiktionaryAPI_FetchRelatedWords_UnsupportedLanguage(t *testing.T) {
	// Setup
	logger := zerolog.New(zerolog.NewTestWriter(t))
	api := &WiktionaryAPI{
		logger:   logger,
		scrapers: make(map[string]wordDomain.DictionaryAPI),
	}

	// Only register French
	mockScraper := new(MockDictionaryAPI)
	api.scrapers["fr"] = mockScraper

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
	mockScraper.AssertNotCalled(t, "FetchRelatedWords")
}

func TestWiktionaryAPI_FetchSuggestions_SupportedLanguage(t *testing.T) {
	// Setup
	logger := zerolog.New(zerolog.NewTestWriter(t))
	api := &WiktionaryAPI{
		logger:   logger,
		scrapers: make(map[string]wordDomain.DictionaryAPI),
	}

	// Create mock scraper
	mockScraper := new(MockDictionaryAPI)
	api.scrapers["fr"] = mockScraper

	// Setup expectations
	expectedSuggestions := []*wordDomain.Word{
		wordDomain.NewWord("test1", "fr"),
		wordDomain.NewWord("test2", "fr"),
	}
	mockScraper.On("FetchSuggestions", mock.Anything, "test", "fr").Return(expectedSuggestions, nil)

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
	api := &WiktionaryAPI{
		logger:   logger,
		scrapers: make(map[string]wordDomain.DictionaryAPI),
	}

	// Only register French
	mockScraper := new(MockDictionaryAPI)
	api.scrapers["fr"] = mockScraper

	// Execute
	ctx := context.Background()
	result, err := api.FetchSuggestions(ctx, "test", "en")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, wordDomain.ErrWordNotFound))
	assert.Contains(t, err.Error(), "unsupported language en")
	mockScraper.AssertNotCalled(t, "FetchSuggestions")
}
