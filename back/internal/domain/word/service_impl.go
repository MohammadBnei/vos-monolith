package word

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Ensure service implements Service interface
var _ Service = (*service)(nil)

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

// Search finds a word by text and language, fetching from external API if needed
func (s *service) Search(ctx context.Context, text, language string) (*Word, error) {
	s.logger.Debug().Str("text", text).Str("language", language).Msg("Searching for word")

	if text == "" {
		s.logger.Warn().Msg("Empty word text provided")
		return nil, ErrInvalidWord
	}

	// Normalize the input text (lowercase, trim spaces)
	normalizedText := strings.ToLower(strings.TrimSpace(text))

	// First, try to find the word in the repository
	word, err := s.repo.FindByText(ctx, normalizedText, language)
	if err == nil {
		s.logger.Debug().Str("text", normalizedText).Msg("Word found in repository")
		return word, nil
	}

	// If not found by exact match, try to find by any form
	if errors.Is(err, ErrWordNotFound) {
		word, err = s.repo.FindByAnyForm(ctx, normalizedText, language)
		if err == nil {
			s.logger.Debug().Str("text", normalizedText).Msg("Word found in repository by form")
			return word, nil
		}
	}

	// If not found in repository, fetch from external API
	if errors.Is(err, ErrWordNotFound) {
		s.logger.Debug().Str("text", normalizedText).Msg("Word not found in repository, fetching from API")
		word, err = s.dictAPI.FetchWord(ctx, normalizedText, language)
		if err != nil {
			s.logger.Error().Err(err).Str("text", normalizedText).Msg("Failed to fetch word from API")
			return nil, fmt.Errorf("failed to fetch word: %w", err)
		}

		// Save the fetched word to the repository
		if err := s.repo.Save(ctx, word); err != nil {
			s.logger.Error().Err(err).Str("text", normalizedText).Msg("Failed to save word to repository")
			// Don't return error here, we still want to return the word to the user
		}

		return word, nil
	}

	// If there was a different error with the repository
	s.logger.Error().Err(err).Str("text", normalizedText).Msg("Repository error")
	return nil, fmt.Errorf("repository error: %w", err)
}

// GetRecentWords retrieves recently searched words
func (s *service) GetRecentWords(ctx context.Context, language string, limit int) ([]*Word, error) {
	s.logger.Debug().Str("language", language).Int("limit", limit).Msg("Getting recent words")

	// Create filter for the repository
	filter := map[string]interface{}{
		"language": language,
	}

	// Get words from repository, sorted by updated_at desc
	words, err := s.repo.List(ctx, filter, limit, 0)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get recent words")
		return nil, fmt.Errorf("failed to get recent words: %w", err)
	}

	return words, nil
}

// GetRelatedWords finds words related to the given word (synonyms, antonyms)
func (s *service) GetRelatedWords(ctx context.Context, wordID string) (*RelatedWords, error) {
	s.logger.Debug().Str("wordID", wordID).Msg("Getting related words")

	// Find the source word
	word, err := s.repo.FindByID(ctx, wordID)
	if err != nil {
		s.logger.Error().Err(err).Str("wordID", wordID).Msg("Failed to find word")
		return nil, fmt.Errorf("failed to find word: %w", err)
	}

	// Initialize result
	result := &RelatedWords{
		SourceWord: word,
		Synonyms:   []*Word{},
		Antonyms:   []*Word{},
	}

	// If the word has no synonyms or antonyms, fetch them
	if len(word.Synonyms) == 0 && len(word.Antonyms) == 0 {
		s.logger.Debug().Str("wordID", wordID).Msg("Fetching related words from API")

		// Fetch related words from API
		relatedWords, err := s.dictAPI.FetchRelatedWords(ctx, word)
		if err != nil {
			s.logger.Error().Err(err).Str("wordID", wordID).Msg("Failed to fetch related words")
			// Return what we have so far, don't fail completely
			return result, nil
		}

		// Update the source word with the new related words
		word.Synonyms = relatedWords.SourceWord.Synonyms
		word.Antonyms = relatedWords.SourceWord.Antonyms
		word.UpdatedAt = time.Now()

		// Save the updated word
		if err := s.repo.Save(ctx, word); err != nil {
			s.logger.Error().Err(err).Str("wordID", wordID).Msg("Failed to save updated word")
			// Don't fail completely, continue with what we have
		}

		// Return the related words
		return relatedWords, nil
	}

	// If the word already has synonyms or antonyms, fetch them from the repository
	for _, syn := range word.Synonyms {
		synWord, err := s.repo.FindByText(ctx, syn, word.Language)
		if err == nil {
			result.Synonyms = append(result.Synonyms, synWord)
		} else {
			// If not found in repository, create a placeholder
			result.Synonyms = append(result.Synonyms, &Word{
				ID:       uuid.New().String(),
				Text:     syn,
				Language: word.Language,
			})
		}
	}

	for _, ant := range word.Antonyms {
		antWord, err := s.repo.FindByText(ctx, ant, word.Language)
		if err == nil {
			result.Antonyms = append(result.Antonyms, antWord)
		} else {
			// If not found in repository, create a placeholder
			result.Antonyms = append(result.Antonyms, &Word{
				ID:       uuid.New().String(),
				Text:     ant,
				Language: word.Language,
			})
		}
	}

	return result, nil
}

// AutoComplete provides autocomplete suggestions based on prefix and language
func (s *service) AutoComplete(ctx context.Context, prefix, language string) ([]string, error) {
	s.logger.Debug().Str("prefix", prefix).Str("language", language).Msg("Getting autocomplete suggestions")

	return s.GetSuggestions(ctx, prefix, language)
}

// GetSuggestions retrieves word suggestions based on a prefix
func (s *service) GetSuggestions(ctx context.Context, prefix, language string) ([]string, error) {
	s.logger.Debug().Str("prefix", prefix).Str("language", language).Msg("Getting word suggestions")

	// Normalize the prefix
	normalizedPrefix := strings.ToLower(strings.TrimSpace(prefix))
	if normalizedPrefix == "" {
		return []string{}, nil
	}

	// First try to get suggestions from the repository
	suggestions, err := s.repo.FindSuggestions(ctx, normalizedPrefix, language, 10)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get suggestions from repository")
		// Continue to try the API
	} else if len(suggestions) > 0 {
		return suggestions, nil
	}

	// If no suggestions from repository, try the API
	suggestions, err = s.dictAPI.FetchSuggestions(ctx, normalizedPrefix, language)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to fetch suggestions from API")
		return nil, fmt.Errorf("failed to fetch suggestions: %w", err)
	}

	return suggestions, nil
}
