package dictionary

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	"voconsteroid/internal/infrastructure/dictionary/acl"
	wordDomain "voconsteroid/internal/domain/word"
)

// WiktionaryAPI implements the word.DictionaryAPI interface for Wiktionary
// It acts as a router to language-specific scrapers and uses the ACL pattern
type WiktionaryAPI struct {
	logger       zerolog.Logger
	// Map of language-specific scrapers
	scrapers     map[string]acl.WiktionaryScraper
	// Adapter for transforming data
	adapter      *acl.WiktionaryAdapter
	// Lookup services for each language
	lookupServices map[string]*acl.LookupService
}

// NewWiktionaryAPI creates a new Wiktionary API router
func NewWiktionaryAPI(logger zerolog.Logger) *WiktionaryAPI {
	baseLogger := logger.With().Str("component", "wiktionary_api").Logger()

	// Create the adapter
	adapter := acl.NewWiktionaryAdapter(baseLogger)

	// Create the API instance
	api := &WiktionaryAPI{
		logger:         baseLogger,
		scrapers:       make(map[string]acl.WiktionaryScraper),
		adapter:        adapter,
		lookupServices: make(map[string]*acl.LookupService),
	}

	// Register language-specific scrapers
	frScraper := NewFrenchWiktionaryScraper(baseLogger)
	api.scrapers["fr"] = frScraper
	
	// Create lookup services for each language
	api.lookupServices["fr"] = acl.NewLookupService(adapter, frScraper, baseLogger)
	
	// Add more language scrapers as they are implemented
	// api.scrapers["en"] = NewEnglishWiktionaryScraper(baseLogger)
	// api.lookupServices["en"] = acl.NewLookupService(adapter, api.scrapers["en"], baseLogger)

	return api
}

// FetchWord routes the request to the appropriate language-specific lookup service
func (w *WiktionaryAPI) FetchWord(ctx context.Context, text, language string) (*wordDomain.Word, error) {
	w.logger.Debug().Str("text", text).Str("language", language).Msg("Routing word fetch request")

	// Get the language-specific lookup service
	lookupService, exists := w.lookupServices[language]
	if !exists {
		w.logger.Warn().Str("language", language).Msg("No specific lookup service for language")
		return nil, fmt.Errorf("unsupported language %s: %w", language, wordDomain.ErrWordNotFound)
	}

	// Delegate to the lookup service
	return lookupService.GetWord(ctx, text, language)
}

// FetchRelatedWords routes the request to the appropriate language-specific lookup service
func (w *WiktionaryAPI) FetchRelatedWords(ctx context.Context, word *wordDomain.Word) (*wordDomain.RelatedWords, error) {
	w.logger.Debug().Str("word", word.Text).Str("language", word.Language).Msg("Routing related words fetch request")

	// Get the language-specific lookup service
	lookupService, exists := w.lookupServices[word.Language]
	if !exists {
		w.logger.Warn().Str("language", word.Language).Msg("No specific lookup service for language")
		return nil, fmt.Errorf("unsupported language %s: %w", word.Language, wordDomain.ErrWordNotFound)
	}

	// Delegate to the lookup service
	return lookupService.GetRelatedWords(ctx, word)
}

// FetchSuggestions routes the request to the appropriate language-specific lookup service
func (w *WiktionaryAPI) FetchSuggestions(ctx context.Context, prefix, language string) ([]string, error) {
	w.logger.Debug().Str("prefix", prefix).Str("language", language).Msg("Routing suggestions fetch request")

	// Get the language-specific lookup service
	lookupService, exists := w.lookupServices[language]
	if !exists {
		w.logger.Warn().Str("language", language).Msg("No specific lookup service for language")
		return nil, fmt.Errorf("unsupported language %s: %w", language, wordDomain.ErrWordNotFound)
	}

	// Delegate to the lookup service
	return lookupService.GetSuggestions(ctx, prefix, language)
}
