package word

import (
	"context"
	"errors"
	
	"github.com/rs/zerolog"
)

var (
	// ErrWordNotFound is returned when a word is not found
	ErrWordNotFound = errors.New("word not found")
	
	// ErrInvalidWord is returned when a word is invalid
	ErrInvalidWord = errors.New("invalid word")
)

// Service defines the interface for word business logic
type Service interface {
	// Search finds a word by text and language, fetching from external API if needed
	Search(ctx context.Context, text, language string) (*Word, error)
	
	// GetRecentWords retrieves recently searched words
	GetRecentWords(ctx context.Context, language string, limit int) ([]*Word, error)
}

// service implements the Service interface
type service struct {
	repo      Repository
	dictAPI   DictionaryAPI
	logger    zerolog.Logger
}

// NewService creates a new word service
func NewService(repo Repository, dictAPI DictionaryAPI, logger zerolog.Logger) Service {
	return &service{
		repo:      repo,
		dictAPI:   dictAPI,
		logger:    logger.With().Str("component", "word_service").Logger(),
	}
}

// Search finds a word by text and language
func (s *service) Search(ctx context.Context, text, language string) (*Word, error) {
	s.logger.Debug().Str("text", text).Str("language", language).Msg("Searching for word")
	
	if text == "" {
		s.logger.Warn().Msg("Empty word text provided")
		return nil, ErrInvalidWord
	}
	
	// Try to find in repository first
	word, err := s.repo.FindByText(ctx, text, language)
	if err == nil {
		s.logger.Debug().Str("text", text).Str("language", language).Msg("Word found in repository")
		return word, nil
	}
	
	s.logger.Debug().Str("text", text).Str("language", language).Err(err).Msg("Word not found in repository, fetching from API")
	
	// If not found, fetch from external API
	word, err = s.dictAPI.FetchWord(ctx, text, language)
	if err != nil {
		s.logger.Error().Err(err).Str("text", text).Str("language", language).Msg("Failed to fetch word from API")
		return nil, err
	}
	
	// Save to repository for future use
	if err := s.repo.Save(ctx, word); err != nil {
		// Log error but don't fail the request
		s.logger.Error().Err(err).Str("word", text).Str("language", language).Msg("Failed to save word to repository")
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
