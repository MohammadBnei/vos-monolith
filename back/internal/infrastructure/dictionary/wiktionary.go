package dictionary

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	wordDomain "voconsteroid/internal/domain/word"
)

// WiktionaryAPI implements the word.DictionaryAPI interface for Wiktionary
// It acts as a router to language-specific scrapers
type WiktionaryAPI struct {
	logger zerolog.Logger
	// Map of language-specific scrapers
	scrapers map[string]wordDomain.DictionaryAPI
}

// NewWiktionaryAPI creates a new Wiktionary API router
func NewWiktionaryAPI(logger zerolog.Logger) *WiktionaryAPI {
	baseLogger := logger.With().Str("component", "wiktionary_api").Logger()

	// Create the API instance
	api := &WiktionaryAPI{
		logger:   baseLogger,
		scrapers: make(map[string]wordDomain.DictionaryAPI),
	}

	// Register language-specific scrapers
	api.scrapers["fr"] = NewFrenchWiktionaryAPI(baseLogger)
	// Add more language scrapers as they are implemented
	// api.scrapers["en"] = NewEnglishWiktionaryAPI(baseLogger)
	// api.scrapers["es"] = NewSpanishWiktionaryAPI(baseLogger)

	return api
}

// FetchWord routes the request to the appropriate language-specific scraper
func (w *WiktionaryAPI) FetchWord(ctx context.Context, text, language string) (*wordDomain.Word, error) {
	w.logger.Debug().Str("text", text).Str("language", language).Msg("Routing word fetch request")

	// Get the language-specific scraper
	scraper, exists := w.scrapers[language]
	if !exists {
		w.logger.Warn().Str("language", language).Msg("No specific scraper for language, using fallback")
		// For now, return an error when no language-specific scraper exists
		return nil, fmt.Errorf("unsupported language %s: %w", language, wordDomain.ErrWordNotFound)

		// Later, we could implement a fallback scraper:
		// return w.fallbackScraper.FetchWord(ctx, text, language)
	}

	// Delegate to the language-specific scraper
	return scraper.FetchWord(ctx, text, language)
}

// FetchRelatedWords routes the request to the appropriate language-specific scraper
func (w *WiktionaryAPI) FetchRelatedWords(ctx context.Context, word *wordDomain.Word) (*wordDomain.RelatedWords, error) {
	w.logger.Debug().Str("word", word.Text).Str("language", word.Language).Msg("Routing related words fetch request")

	// Get the language-specific scraper
	scraper, exists := w.scrapers[word.Language]
	if !exists {
		w.logger.Warn().Str("language", word.Language).Msg("No specific scraper for language, using fallback")
		// For now, return an error when no language-specific scraper exists
		return nil, fmt.Errorf("unsupported language %s: %w", word.Language, wordDomain.ErrWordNotFound)
	}

	// Delegate to the language-specific scraper
	return scraper.FetchRelatedWords(ctx, word)
}

// FetchSuggestions routes the request to the appropriate language-specific scraper
func (w *WiktionaryAPI) FetchSuggestions(ctx context.Context, prefix, language string) ([]string, error) {
	w.logger.Debug().Str("prefix", prefix).Str("language", language).Msg("Routing suggestions fetch request")

	// Get the language-specific scraper
	scraper, exists := w.scrapers[language]
	if !exists {
		w.logger.Warn().Str("language", language).Msg("No specific scraper for language, using fallback")
		// For now, return an error when no language-specific scraper exists
		return nil, fmt.Errorf("unsupported language %s: %w", language, wordDomain.ErrWordNotFound)
	}

	// Delegate to the language-specific scraper
	return scraper.FetchSuggestions(ctx, prefix, language)
}
