package word

import (
	"context"
	"errors"
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
}

// NewService creates a new word service
func NewService(repo Repository, dictAPI DictionaryAPI) Service {
	return &service{
		repo:      repo,
		dictAPI:   dictAPI,
	}
}

// Search finds a word by text and language
func (s *service) Search(ctx context.Context, text, language string) (*Word, error) {
	if text == "" {
		return nil, ErrInvalidWord
	}
	
	// Try to find in repository first
	word, err := s.repo.FindByText(ctx, text, language)
	if err == nil {
		return word, nil
	}
	
	// If not found, fetch from external API
	word, err = s.dictAPI.FetchWord(ctx, text, language)
	if err != nil {
		return nil, err
	}
	
	// Save to repository for future use
	if err := s.repo.Save(ctx, word); err != nil {
		// Log error but don't fail the request
		// logger could be injected into the service
		return word, nil
	}
	
	return word, nil
}

// GetRecentWords retrieves recently searched words
func (s *service) GetRecentWords(ctx context.Context, language string, limit int) ([]*Word, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}
	
	filter := map[string]interface{}{
		"language": language,
	}
	
	return s.repo.List(ctx, filter, limit, 0)
}
