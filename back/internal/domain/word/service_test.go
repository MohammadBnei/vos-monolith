package word

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of the Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) FindByText(ctx context.Context, text, language string) (*Word, error) {
	args := m.Called(ctx, text, language)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Word), args.Error(1)
}

func (m *MockRepository) FindByAnyForm(ctx context.Context, text, language string) (*Word, error) {
	args := m.Called(ctx, text, language)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Word), args.Error(1)
}

func (m *MockRepository) Save(ctx context.Context, word *Word) error {
	args := m.Called(ctx, word)
	return args.Error(0)
}

func (m *MockRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*Word, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Word), args.Error(1)
}

func (m *MockRepository) FindByPrefix(ctx context.Context, prefix, language string, limit int) ([]*Word, error) {
	args := m.Called(ctx, prefix, language, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Word), args.Error(1)
}

// MockDictionaryAPI is a mock implementation of the DictionaryAPI interface
type MockDictionaryAPI struct {
	mock.Mock
}

func (m *MockDictionaryAPI) FetchWord(ctx context.Context, text, language string) (*Word, error) {
	args := m.Called(ctx, text, language)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Word), args.Error(1)
}

func (m *MockDictionaryAPI) FetchRelatedWords(ctx context.Context, word *Word) (*RelatedWords, error) {
	args := m.Called(ctx, word)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RelatedWords), args.Error(1)
}

func (m *MockDictionaryAPI) FetchSuggestions(ctx context.Context, prefix, language string) ([]string, error) {
	args := m.Called(ctx, prefix, language)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func TestNewService(t *testing.T) {
	// Setup
	repo := new(MockRepository)
	dictAPI := new(MockDictionaryAPI)
	logger := zerolog.New(zerolog.NewTestWriter(t))

	// Execute
	svc := NewService(repo, dictAPI, logger)

	// Assert
	assert.NotNil(t, svc)
}

// setupTestService creates a service with mocks for testing
func setupTestService(t *testing.T) (*MockRepository, *MockDictionaryAPI, Service) {
	repo := new(MockRepository)
	dictAPI := new(MockDictionaryAPI)
	logger := zerolog.New(zerolog.NewTestWriter(t))

	svc := NewService(repo, dictAPI, logger)

	return repo, dictAPI, svc
}

func TestSearch_ExistingWord(t *testing.T) {
	// Setup
	repo, dictAPI, svc := setupTestService(t)

	ctx := context.Background()
	expectedWord := &Word{
		Text:      "test",
		Language:  "en",
		CreatedAt: time.Now(),
	}

	// Expect repository to find the word by text first
	repo.On("FindByText", ctx, "test", "en").Return(expectedWord, nil)

	// Add expectation for FindByAnyForm (won't be called but needs to be mocked)
	repo.On("FindByAnyForm", ctx, "test", "en").Return(nil, errors.New("not found"))

	// Execute
	word, err := svc.Search(ctx, "test", "en")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedWord, word)
	repo.AssertExpectations(t)
	dictAPI.AssertNotCalled(t, "FetchWord")

	// Verify that FindByAnyForm was not called since we found the word by text
	repo.AssertNotCalled(t, "FindByAnyForm")
}

func TestSearch_NewWord(t *testing.T) {
	// Setup
	repo, dictAPI, svc := setupTestService(t)

	ctx := context.Background()
	expectedWord := &Word{
		Text:      "test",
		Language:  "en",
		CreatedAt: time.Now(),
	}

	// Expect repository to not find the word by text
	repo.On("FindByText", ctx, "test", "en").Return(nil, errors.New("not found"))

	// Expect repository to not find the word by any form
	repo.On("FindByAnyForm", ctx, "test", "en").Return(nil, errors.New("not found"))

	// Expect dictionary API to fetch the word
	dictAPI.On("FetchWord", ctx, "test", "en").Return(expectedWord, nil)

	// Expect repository to save the word
	repo.On("Save", ctx, expectedWord).Return(nil)

	// Execute
	word, err := svc.Search(ctx, "test", "en")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedWord, word)
	repo.AssertExpectations(t)
	dictAPI.AssertExpectations(t)
}

func TestSearch_EmptyText(t *testing.T) {
	// Setup
	repo, dictAPI, svc := setupTestService(t)

	ctx := context.Background()

	// Execute
	word, err := svc.Search(ctx, "", "en")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidWord, err)
	assert.Nil(t, word)
	repo.AssertNotCalled(t, "FindByText")
	repo.AssertNotCalled(t, "FindByAnyForm")
	dictAPI.AssertNotCalled(t, "FetchWord")
}

func TestSearch_APIError(t *testing.T) {
	// Setup
	repo, dictAPI, svc := setupTestService(t)

	ctx := context.Background()
	apiErr := errors.New("API error")

	// Expect repository to not find the word by text
	repo.On("FindByText", ctx, "test", "en").Return(nil, errors.New("not found"))

	// Expect repository to not find the word by any form
	repo.On("FindByAnyForm", ctx, "test", "en").Return(nil, errors.New("not found"))

	// Expect dictionary API to return an error
	dictAPI.On("FetchWord", ctx, "test", "en").Return(nil, apiErr)

	// Execute
	word, err := svc.Search(ctx, "test", "en")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
	assert.Nil(t, word)
	repo.AssertExpectations(t)
	dictAPI.AssertExpectations(t)
}

func TestGetRecentWords(t *testing.T) {
	// Setup
	repo, _, svc := setupTestService(t)

	ctx := context.Background()
	expectedWords := []*Word{
		{Text: "test1", Language: "en"},
		{Text: "test2", Language: "en"},
	}

	// Expect repository to list words
	expectedFilter := map[string]interface{}{"language": "en"}
	repo.On("List", ctx, expectedFilter, 10, 0).Return(expectedWords, nil)

	// Execute
	words, err := svc.GetRecentWords(ctx, "en", 10)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedWords, words)
	repo.AssertExpectations(t)
}

func TestGetRecentWords_DefaultLimit(t *testing.T) {
	// Setup
	repo, _, svc := setupTestService(t)

	ctx := context.Background()
	expectedWords := []*Word{
		{Text: "test1", Language: "en"},
		{Text: "test2", Language: "en"},
	}

	// Expect repository to list words with default limit
	expectedFilter := map[string]interface{}{"language": "en"}
	repo.On("List", ctx, expectedFilter, 10, 0).Return(expectedWords, nil)

	// Execute with zero limit (should use default)
	words, err := svc.GetRecentWords(ctx, "en", 0)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedWords, words)
	repo.AssertExpectations(t)
}

func TestGetRecentWords_RepositoryError(t *testing.T) {
	// Setup
	repo, _, svc := setupTestService(t)

	ctx := context.Background()
	repoErr := errors.New("repository error")

	// Expect repository to return an error
	expectedFilter := map[string]interface{}{"language": "en"}
	repo.On("List", ctx, expectedFilter, 10, 0).Return(nil, repoErr)

	// Execute
	words, err := svc.GetRecentWords(ctx, "en", 10)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, words)
	assert.Equal(t, repoErr, err)
	repo.AssertExpectations(t)
}

func TestAutoComplete(t *testing.T) {
	// Setup
	repo, dictAPI, svc := setupTestService(t)

	ctx := context.Background()

	// Create test data
	localWords := []*Word{
		{Text: "test1", Language: "en"},
		{Text: "test2", Language: "en"},
	}

	apiSuggestions := []string{"test2", "test3"} // Duplicate that should be deduplicated

	// Expect repository to find words by prefix
	repo.On("FindByPrefix", ctx, "test", "en", 5).Return(localWords, nil)

	// Expect dictionary API to fetch suggestions
	dictAPI.On("FetchSuggestions", ctx, "test", "en").Return(apiSuggestions, nil)

	// Execute
	results, err := svc.AutoComplete(ctx, "test", "en")

	// Assert
	assert.NoError(t, err)
	assert.Len(t, results, 3)

	// Check that results are sorted alphabetically
	assert.Equal(t, "test1", results[0])
	assert.Equal(t, "test2", results[1])
	assert.Equal(t, "test3", results[2])

	repo.AssertExpectations(t)
	dictAPI.AssertExpectations(t)
}

func TestAutoComplete_ShortPrefix(t *testing.T) {
	// Setup
	repo, dictAPI, svc := setupTestService(t)

	ctx := context.Background()

	// Execute with prefix that's too short
	results, err := svc.AutoComplete(ctx, "t", "en")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidWord, err)
	assert.Nil(t, results)

	// Verify no calls were made
	repo.AssertNotCalled(t, "FindByPrefix")
	dictAPI.AssertNotCalled(t, "FetchSuggestions")
}

func TestAutoComplete_RepositoryError(t *testing.T) {
	// Setup
	repo, dictAPI, svc := setupTestService(t)

	ctx := context.Background()
	repoErr := errors.New("repository error")

	// Create test data for API
	apiSuggestions := []string{"test3"}

	// Expect repository to return an error
	repo.On("FindByPrefix", ctx, "test", "en", 5).Return(nil, repoErr)

	// API should still be called and succeed
	dictAPI.On("FetchSuggestions", ctx, "test", "en").Return(apiSuggestions, nil)

	// Execute
	results, err := svc.AutoComplete(ctx, "test", "en")

	// Assert
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "test3", results[0])

	repo.AssertExpectations(t)
	dictAPI.AssertExpectations(t)
}

func TestAutoComplete_APIError(t *testing.T) {
	// Setup
	repo, dictAPI, svc := setupTestService(t)

	ctx := context.Background()
	apiErr := errors.New("API error")

	// Create test data for repository
	localWords := []*Word{
		{Text: "test1", Language: "en"},
	}

	// Repository should succeed
	repo.On("FindByPrefix", ctx, "test", "en", 5).Return(localWords, nil)

	// API should fail
	dictAPI.On("FetchSuggestions", ctx, "test", "en").Return([]string(nil), apiErr)

	// Execute
	results, err := svc.AutoComplete(ctx, "test", "en")

	// Assert
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "test1", results[0])

	repo.AssertExpectations(t)
	dictAPI.AssertExpectations(t)
}
