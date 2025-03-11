package word

import (
	"context"
	"sort"
	"strings"

	"github.com/rs/zerolog"
)

// Service defines the interface for word business logic
type Service interface {
	// Search finds a word by text and language, fetching from external API if needed
	Search(ctx context.Context, text, language string) (*Word, error)

	// GetRecentWords retrieves recently searched words
	GetRecentWords(ctx context.Context, language string, limit int) ([]*Word, error)

	// GetRelatedWords finds words related to the given word (synonyms, antonyms)
	GetRelatedWords(ctx context.Context, wordID string) (*RelatedWords, error)

	// AutoComplete provides autocomplete suggestions based on prefix and language
	AutoComplete(ctx context.Context, prefix, language string) ([]string, error)
	
	// GetSuggestions retrieves word suggestions based on a prefix
	GetSuggestions(ctx context.Context, prefix, language string) ([]string, error)
}

// RelatedWords groups words related to a specific word
type RelatedWords struct {
	SourceWord *Word   `json:"source_word"`
	Synonyms   []*Word `json:"synonyms,omitempty"`
	Antonyms   []*Word `json:"antonyms,omitempty"`
}

// service implements the Service interface
type service struct {
	repo    Repository
	dictAPI DictionaryAPI
	logger  zerolog.Logger
}

// NewService creates a new word service
func NewService(repo Repository, dictAPI DictionaryAPI, logger zerolog.Logger) Service {
	return &service{
		repo:    repo,
		dictAPI: dictAPI,
		logger:  logger.With().Str("component", "word_service").Logger(),
	}
}

// Search finds a word by text and language
func (s *service) Search(ctx context.Context, text, language string) (*Word, error) {
	s.logger.Debug().Str("text", text).Str("language", language).Msg("Searching for word")

	if text == "" {
		s.logger.Warn().Msg("Empty word text provided")
		return nil, ErrInvalidWord
	}

	// Normalize the input text (lowercase, trim spaces)
	normalizedText := strings.ToLower(strings.TrimSpace(text))

	// Try to find by any form first (prioritizing user flexibility)
	word, err := s.repo.FindByAnyForm(ctx, normalizedText, language)
	if err == nil {
		s.logger.Debug().Str("text", normalizedText).Str("found_form", word.Text).Msg("Word found in repository by form")
		return word, nil
	}

	// If not found by any form, try exact match as fallback
	word, err = s.repo.FindByText(ctx, normalizedText, language)
	if err == nil {
		s.logger.Debug().Str("text", normalizedText).Str("language", language).Msg("Word found in repository by exact match")
		return word, nil
	}

	s.logger.Debug().Str("text", normalizedText).Str("language", language).Msg("Word not found in repository, fetching from API")

	// If still not found, fetch from external API
	word, err = s.dictAPI.FetchWord(ctx, normalizedText, language)
	if err != nil {
		s.logger.Error().Err(err).Str("text", normalizedText).Str("language", language).Msg("Failed to fetch word from API")
		return nil, err
	}

	// Save to repository for future use
	if err := s.repo.Save(ctx, word); err != nil {
		// Log error but don't fail the request
		s.logger.Error().Err(err).Str("word", normalizedText).Str("language", language).Msg("Failed to save word to repository")
		return word, nil
	}

	return word, nil
}

// GetRecentWords retrieves recently searched words
func (s *service) GetRecentWords(ctx context.Context, language string, limit int) ([]*Word, error) {
	s.logger.Debug().Str("language", language).Int("limit", limit).Msg("Getting recent words")

	if limit <= 0 {
		limit = 10 // Default limit
		s.logger.Debug().Int("limit", limit).Msg("Using default limit")
	}

	filter := map[string]interface{}{
		"language": language,
	}

	words, err := s.repo.List(ctx, filter, limit, 0)
	if err != nil {
		s.logger.Error().Err(err).Str("language", language).Int("limit", limit).Msg("Failed to get recent words")
		return nil, err
	}

	s.logger.Debug().Str("language", language).Int("count", len(words)).Msg("Retrieved recent words")
	return words, nil
}

// GetRelatedWords finds words related to the given word
func (s *service) GetRelatedWords(ctx context.Context, wordID string) (*RelatedWords, error) {
	s.logger.Debug().Str("wordID", wordID).Msg("Getting related words")

	// First, get the source word
	filter := map[string]interface{}{
		"id": wordID,
	}

	words, err := s.repo.List(ctx, filter, 1, 0)
	if err != nil || len(words) == 0 {
		s.logger.Error().Err(err).Str("wordID", wordID).Msg("Failed to get source word")
		return nil, ErrWordNotFound
	}

	sourceWord := words[0]
	result := &RelatedWords{
		SourceWord: sourceWord,
		Synonyms:   []*Word{},
		Antonyms:   []*Word{},
	}

	// Get synonyms
	if len(sourceWord.Synonyms) > 0 {
		for _, syn := range sourceWord.Synonyms {
			// For each synonym, try to find the full word entry
			word, err := s.Search(ctx, syn, sourceWord.Language)
			if err == nil {
				result.Synonyms = append(result.Synonyms, word)
			}
		}
	}

	// Get antonyms
	if len(sourceWord.Antonyms) > 0 {
		for _, ant := range sourceWord.Antonyms {
			// For each antonym, try to find the full word entry
			word, err := s.Search(ctx, ant, sourceWord.Language)
			if err == nil {
				result.Antonyms = append(result.Antonyms, word)
			}
		}
	}

	s.logger.Debug().
		Str("wordID", wordID).
		Int("synonyms", len(result.Synonyms)).
		Int("antonyms", len(result.Antonyms)).
		Msg("Retrieved related words")

	return result, nil
}

func (s *service) AutoComplete(ctx context.Context, prefix, language string) ([]string, error) {
	if len(prefix) < 2 {
		return nil, ErrInvalidWord
	}

	s.logger.Debug().Str("prefix", prefix).Str("language", language).Msg("Getting autocomplete suggestions")

	// Get local results (fail silently to continue with API results if DB fails)
	localResults, _ := s.repo.FindByPrefix(ctx, prefix, language, 5)

	// Get external suggestions (fail silently to continue with local results if API fails)
	apiResults, _ := s.dictAPI.FetchSuggestions(ctx, prefix, language)

	// Merge results and sort by relevance
	dedupMap := make(map[string]struct{})
	for _, word := range localResults {
		dedupMap[word.Text] = struct{}{}
	}
	for _, word := range apiResults {
		dedupMap[word] = struct{}{}

	}

	finalResults := make([]string, 0, len(dedupMap))
	for word := range dedupMap {
		finalResults = append(finalResults, word)
	}

	// Sort alphabetically as a baseline
	sort.Slice(finalResults, func(i, j int) bool {
		return finalResults[i] < finalResults[j]
	})

	s.logger.Debug().Int("count", len(finalResults)).Msg("Returning autocomplete suggestions")
	return finalResults, nil
}
