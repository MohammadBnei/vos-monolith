package dictionary

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/rs/zerolog"

	"voconsteroid/internal/domain/word"
)

// WiktionaryAPI implements the word.DictionaryAPI interface for Wiktionary
type WiktionaryAPI struct {
	baseURL string
	logger  zerolog.Logger
}

// NewWiktionaryAPI creates a new Wiktionary scraper
func NewWiktionaryAPI(logger zerolog.Logger) *WiktionaryAPI {
	return &WiktionaryAPI{
		baseURL: "https://en.wiktionary.org/wiki",
		logger:  logger.With().Str("component", "wiktionary_scraper").Logger(),
	}
}

// FetchWord retrieves word information from Wiktionary by scraping the web page
func (w *WiktionaryAPI) FetchWord(ctx context.Context, text, language string) (*word.Word, error) {
	w.logger.Debug().Str("text", text).Str("language", language).Msg("Fetching word from Wiktionary")
	
	// Create a new collector
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		colly.MaxDepth(1),
	)

	// Set timeout
	c.SetRequestTimeout(10 * time.Second)

	// Create a new word
	newWord := word.NewWord(text, language)
	
	// Track if we found any definitions
	foundDefinitions := false
	
	// Extract definitions
	c.OnHTML("ol li", func(e *colly.HTMLElement) {
		// Check if this is a definition list (usually under the language section)
		if e.DOM.ParentsFiltered("div[id^='"+language+"']").Length() > 0 || 
		   e.DOM.ParentsFiltered("h2:contains('"+strings.Title(language)+"')").Length() > 0 {
			definition := strings.TrimSpace(e.Text)
			if definition != "" {
				newWord.Definitions = append(newWord.Definitions, definition)
				foundDefinitions = true
			}
		}
	})

	// Extract examples
	c.OnHTML("div.example-needed, ul.citations li, div.citation-whole", func(e *colly.HTMLElement) {
		if e.DOM.ParentsFiltered("div[id^='"+language+"']").Length() > 0 || 
		   e.DOM.ParentsFiltered("h2:contains('"+strings.Title(language)+"')").Length() > 0 {
			example := strings.TrimSpace(e.Text)
			if example != "" {
				newWord.Examples = append(newWord.Examples, example)
			}
		}
	})

	// Extract pronunciation
	c.OnHTML("span.IPA", func(e *colly.HTMLElement) {
		if e.DOM.ParentsFiltered("div[id^='"+language+"']").Length() > 0 || 
		   e.DOM.ParentsFiltered("h2:contains('"+strings.Title(language)+"')").Length() > 0 {
			if newWord.Pronunciation == "" {
				newWord.Pronunciation = strings.TrimSpace(e.Text)
			}
		}
	})

	// Extract translations
	c.OnHTML("div.translations li", func(e *colly.HTMLElement) {
		if e.DOM.ParentsFiltered("div[id^='"+language+"']").Length() > 0 || 
		   e.DOM.ParentsFiltered("h2:contains('"+strings.Title(language)+"')").Length() > 0 {
			langCode := e.ChildAttr("span.language", "lang")
			translation := strings.TrimSpace(e.ChildText("span.translation"))
			if langCode != "" && translation != "" {
				newWord.Translations[langCode] = translation
			}
		}
	})

	// Extract etymology if available
	c.OnHTML("div#Etymology, div[id^='Etymology_']", func(e *colly.HTMLElement) {
		if e.DOM.ParentsFiltered("div[id^='"+language+"']").Length() > 0 || 
		   e.DOM.ParentsFiltered("h2:contains('"+strings.Title(language)+"')").Length() > 0 {
			etymology := strings.TrimSpace(e.Text)
			if etymology != "" {
				newWord.Etymology = etymology
			}
		}
	})

	// Handle errors
	c.OnError(func(r *colly.Response, err error) {
		w.logger.Error().Err(err).Str("url", r.Request.URL.String()).Msg("Failed to scrape page")
	})

	// Build URL for the web page
	url := fmt.Sprintf("%s/%s", w.baseURL, text)
	
	// Check if context is done
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Visit the page
		err := c.Visit(url)
		if err != nil {
			w.logger.Error().Err(err).Str("url", url).Msg("Failed to visit page")
			return nil, fmt.Errorf("failed to visit page: %w", err)
		}
	}

	// Wait until scraping is finished
	c.Wait()

	// If no definitions were found, return an error
	if !foundDefinitions || len(newWord.Definitions) == 0 {
		w.logger.Warn().Str("text", text).Str("language", language).Msg("No word data found")
		return nil, fmt.Errorf("no word data found: %w", word.ErrWordNotFound)
	}

	w.logger.Debug().
		Str("text", text).
		Str("language", language).
		Int("definitions", len(newWord.Definitions)).
		Int("examples", len(newWord.Examples)).
		Int("translations", len(newWord.Translations)).
		Msg("Successfully fetched word data")
		
	return newWord, nil
}
