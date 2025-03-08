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

// FrenchWiktionaryAPI implements the word.DictionaryAPI interface for French Wiktionary
type FrenchWiktionaryAPI struct {
	logger     zerolog.Logger
	getBaseURL func(language string) string
}

// NewFrenchWiktionaryAPI creates a new French Wiktionary scraper
func NewFrenchWiktionaryAPI(logger zerolog.Logger) *FrenchWiktionaryAPI {
	return &FrenchWiktionaryAPI{
		logger: logger.With().Str("component", "fr_wiktionary_scraper").Logger(),
		getBaseURL: func(language string) string {
			return "https://fr.wiktionary.org/wiki"
		},
	}
}

// FetchWord retrieves word information from French Wiktionary by scraping the web page
func (w *FrenchWiktionaryAPI) FetchWord(ctx context.Context, text, language string) (*word.Word, error) {
	w.logger.Debug().Str("text", text).Str("language", language).Msg("Fetching word from French Wiktionary")

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

	// Track the current section
	var currentSection string
	var inDefinitionList bool

	// Extract etymology
	c.OnHTML("div.mw-heading-3:has(span.titreetym) + dl", func(e *colly.HTMLElement) {
		etymology := strings.TrimSpace(e.Text)
		if etymology != "" {
			newWord.Etymology = etymology
		}
	})

	// Extract plural form from the flextable
	c.OnHTML("table.flextable-fr-mfsp", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(_ int, tr *colly.HTMLElement) {
			if strings.Contains(tr.Text, "Pluriel") {
				tr.ForEach("td", func(_ int, td *colly.HTMLElement) {
					plural := strings.TrimSpace(td.Text)
					if plural != "" && plural != newWord.Text {
						// Store plural in translations as a special case
						newWord.Translations["plural"] = plural
					}
				})
			}
		})
	})

	// Extract pronunciation
	c.OnHTML("span.API", func(e *colly.HTMLElement) {
		if newWord.Pronunciation == "" {
			pronunciation := strings.TrimSpace(e.Text)
			if strings.HasPrefix(pronunciation, "\\") && strings.HasSuffix(pronunciation, "\\") {
				newWord.Pronunciation = pronunciation
			}
		}
	})

	// Track sections to know when we're in definitions, synonyms, etc.
	c.OnHTML("div.mw-heading-3", func(e *colly.HTMLElement) {
		if e.ChildText("span.titredef") != "" {
			currentSection = "definitions"
			inDefinitionList = false
		} else if e.ChildText("span.titresyno") != "" {
			currentSection = "synonyms"
			inDefinitionList = false
		} else {
			currentSection = ""
			inDefinitionList = false
		}
	})

	// Extract definitions and examples
	c.OnHTML("ol", func(e *colly.HTMLElement) {
		if currentSection == "definitions" && !inDefinitionList {
			inDefinitionList = true
			e.ForEach("li", func(i int, li *colly.HTMLElement) {
				// Get the main definition text (excluding nested elements)
				definitionText := strings.TrimSpace(li.Text)

				// Skip empty definitions
				if definitionText == "" {
					return
				}

				// Add the definition
				newWord.Definitions = append(newWord.Definitions, definitionText)

				// Look for examples within this definition
				li.ForEach("ul li span.example", func(_ int, example *colly.HTMLElement) {
					exampleText := strings.TrimSpace(example.Text)
					if exampleText != "" {
						// Clean up the example text
						exampleText = strings.ReplaceAll(exampleText, "« ", "")
						exampleText = strings.ReplaceAll(exampleText, " »", "")
						newWord.Examples = append(newWord.Examples, exampleText)
					}
				})

				foundDefinitions = true
			})
		}
	})

	// Extract synonyms
	c.OnHTML("div.mw-heading-4:has(span.titresyno) + ul", func(e *colly.HTMLElement) {
		e.ForEach("li", func(_ int, li *colly.HTMLElement) {
			synonym := strings.TrimSpace(li.Text)
			if synonym != "" {
				newWord.Synonyms = append(newWord.Synonyms, synonym)
			}
		})
	})

	// Build URL for the web page
	baseURL := w.getBaseURL(language)
	url := fmt.Sprintf("%s/%s", baseURL, text)

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
		Int("synonyms", len(newWord.Synonyms)).
		Str("etymology", newWord.Etymology).
		Str("pronunciation", newWord.Pronunciation).
		Msg("Successfully fetched word data from French Wiktionary")

	return newWord, nil
}
