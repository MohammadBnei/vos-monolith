package acl

import (
	"context"
	"fmt"
	
	"github.com/rs/zerolog"
	
	wordDomain "voconsteroid/internal/domain/word"
)

// LookupService provides methods for looking up words
type LookupService struct {
	adapter *WiktionaryAdapter
	scraper WiktionaryScraper
	logger  zerolog.Logger
}

// NewLookupService creates a new LookupService
func NewLookupService(adapter *WiktionaryAdapter, scraper WiktionaryScraper, logger zerolog.Logger) *LookupService {
	return &LookupService{
		adapter: adapter,
		scraper: scraper,
		logger:  logger.With().Str("component", "lookup_service").Logger(),
	}
}

// GetWord retrieves a word from Wiktionary
func (s *LookupService) GetWord(ctx context.Context, text, language string) (*wordDomain.Word, error) {
	s.logger.Debug().Str("text", text).Str("language", language).Msg("Looking up word")
	
	return s.adapter.FetchAndTransformWord(ctx, s.scraper, text, language)
}

// GetRelatedWords retrieves related words for a word
func (s *LookupService) GetRelatedWords(ctx context.Context, word *wordDomain.Word) (*wordDomain.RelatedWords, error) {
	s.logger.Debug().Str("word", word.Text).Str("language", word.Language).Msg("Looking up related words")
	
	return s.adapter.FetchAndTransformRelatedWords(ctx, s.scraper, word)
}

// GetSuggestions retrieves suggestions for a prefix
func (s *LookupService) GetSuggestions(ctx context.Context, prefix, language string) ([]string, error) {
	s.logger.Debug().Str("prefix", prefix).Str("language", language).Msg("Looking up suggestions")
	
	return s.scraper.FetchSuggestionsData(ctx, prefix, language)
}

// EnrichMissingFields enriches a word with missing fields
func (s *LookupService) EnrichMissingFields(ctx context.Context, word *wordDomain.Word, status EnrichmentStatus) (*wordDomain.Word, error) {
	s.logger.Debug().Str("word", word.Text).Str("language", word.Language).Msg("Enriching word with missing fields")
	
	if word == nil {
		return nil, fmt.Errorf("cannot enrich nil word")
	}
	
	return s.adapter.EnrichWord(ctx, s.scraper, word, status)
}
