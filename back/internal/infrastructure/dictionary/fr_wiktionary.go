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
	var currentDefinitionIndex int

	// Extract etymology
	c.OnHTML("div.mw-heading-3:has(span.titreetym) + dl", func(e *colly.HTMLElement) {
		etymology := strings.TrimSpace(e.Text)
		if etymology != "" {
			w.logger.Debug().Str("etymology", etymology).Msg("Found etymology")
			newWord.Etymology = etymology
		}
	})

	// Debug HTML structure
	c.OnHTML("html", func(e *colly.HTMLElement) {
		w.logger.Debug().Msg("Page loaded successfully")
	})

	// Extract word type (masculine/feminine) and plural form from the flextable
	c.OnHTML("table.flextable-fr-mfsp, p span.ligne-de-forme", func(e *colly.HTMLElement) {
		// Check if this is a table with plural information
		if e.Name == "table" {
			e.ForEach("tr", func(_ int, tr *colly.HTMLElement) {
				if strings.Contains(tr.Text, "Pluriel") {
					tr.ForEach("td", func(_ int, td *colly.HTMLElement) {
						plural := strings.TrimSpace(td.Text)
						if plural != "" && plural != newWord.Text {
							// Store plural in translations as a special case
							w.logger.Debug().Str("plural", plural).Msg("Found plural form")
							newWord.Translations["plural"] = plural
						}
					})
				}
			})
		} else if e.Name == "span" {
			// Extract word type (masculin/féminin)
			wordType := strings.TrimSpace(e.Text)
			if strings.Contains(wordType, "masculin") || strings.Contains(wordType, "féminin") {
				w.logger.Debug().Str("wordType", wordType).Msg("Found word type")
				newWord.Gender = wordType
			}
		}
	})

	// Extract pronunciation
	c.OnHTML("span.API", func(e *colly.HTMLElement) {
		if newWord.Pronunciation == "" {
			pronunciation := strings.TrimSpace(e.Text)
			if strings.HasPrefix(pronunciation, "\\") && strings.HasSuffix(pronunciation, "\\") {
				w.logger.Debug().Str("pronunciation", pronunciation).Msg("Found pronunciation")
				newWord.Pronunciation = pronunciation
			}
		}
	})

	// Track sections to know when we're in definitions, synonyms, etc.
	// Handle both div.mw-heading-3 and h3 elements that might contain section headings
	c.OnHTML("div.mw-heading-3, h3", func(e *colly.HTMLElement) {
		titleDef := e.ChildText("span.titredef")
		if titleDef != "" {
			w.logger.Debug().Str("section", titleDef).Msg("Found definition section")
			currentSection = "definitions"
			inDefinitionList = false
			currentDefinitionIndex = 0
		} else if e.ChildText("span.titresyno") != "" {
			currentSection = "synonyms"
			inDefinitionList = false
		} else {
			// Don't reset the section if we're not recognizing a new one
			// This helps maintain context between different HTML elements
		}
	})

	// Track definition type (Botanique, Cuisine, etc.)
	var currentDefinitionType string

	// Extract definition type from parentheses
	c.OnHTML("li span.emploi", func(e *colly.HTMLElement) {
		defType := strings.TrimSpace(e.Text)
		// Clean up the definition type
		defType = strings.TrimPrefix(defType, "(")
		defType = strings.TrimSuffix(defType, ")")
		if defType != "" {
			currentDefinitionType = defType
			w.logger.Debug().Str("definitionType", defType).Msg("Found definition type")
		}
	})

	// Extract definitions from any ordered list following a definition heading
	// This handles various HTML structures that might contain definitions
	c.OnHTML("div.mw-heading-3:has(span.titredef) + p + ol, div.mw-heading-3:has(span.titredef) + ol, h3:has(span.titredef) + ol, h3:has(span.titredef) + p + ol", func(e *colly.HTMLElement) {
		w.logger.Debug().Msg("Found definition list")
		inDefinitionList = true
		
		// Process each list item as a definition
		e.ForEach("li", func(i int, li *colly.HTMLElement) {
			// Check for definition type within this list item
			defType := li.ChildText("span.emploi")
			if defType != "" {
				// Clean up the definition type
				defType = strings.TrimPrefix(defType, "(")
				defType = strings.TrimSuffix(defType, ")")
				currentDefinitionType = defType
				w.logger.Debug().Str("definitionType", defType).Msg("Found inline definition type")
			}
			
			// Get the main definition text without examples
			var definitionText string
			
			// First, try to get just the direct text content without child elements
			definitionText = li.DOM.Contents().Not("ul").Not("span.example").Text()
			definitionText = strings.TrimSpace(definitionText)
			
			// If that fails, fall back to getting all text and cleaning it later
			if definitionText == "" {
				definitionText = strings.TrimSpace(li.Text)
			}
			
			// Skip empty definitions
			if definitionText == "" {
				return
			}
			
			// Add definition type prefix if available
			if currentDefinitionType != "" {
				definitionText = "(" + currentDefinitionType + ") " + definitionText
			}
			
			// Add the definition
			newWord.Definitions = append(newWord.Definitions, definitionText)
			currentDefinitionIndex = len(newWord.Definitions) - 1
			w.logger.Debug().Int("index", currentDefinitionIndex).Str("definition", definitionText).Msg("Found definition")
			
			// Extract examples separately
			li.ForEach("span.example", func(_ int, example *colly.HTMLElement) {
				exampleText := strings.TrimSpace(example.Text)
				if exampleText != "" {
					// Clean up the example text
					exampleText = strings.ReplaceAll(exampleText, "« ", "")
					exampleText = strings.ReplaceAll(exampleText, " »", "")
					w.logger.Debug().Str("example", exampleText).Msg("Found example")
					newWord.Examples = append(newWord.Examples, exampleText)
				}
			})
			
			foundDefinitions = true
		})
	})
	
	// Backup method to extract definitions if the above selector doesn't work
	c.OnHTML("ol", func(e *colly.HTMLElement) {
		// Only process if we haven't found definitions yet and we're in a definition section
		if !foundDefinitions && currentSection == "definitions" && !inDefinitionList {
			w.logger.Debug().Msg("Using backup method to find definitions")
			inDefinitionList = true
			
			e.ForEach("li", func(i int, li *colly.HTMLElement) {
				// Check for definition type within this list item
				defType := li.ChildText("span.emploi")
				if defType != "" {
					// Clean up the definition type
					defType = strings.TrimPrefix(defType, "(")
					defType = strings.TrimSuffix(defType, ")")
					currentDefinitionType = defType
					w.logger.Debug().Str("definitionType", defType).Msg("Found backup definition type")
				}
				
				// Get the main definition text without examples
				var definitionText string
				
				// First, try to get just the direct text content without child elements
				definitionText = li.DOM.Contents().Not("ul").Not("span.example").Text()
				definitionText = strings.TrimSpace(definitionText)
				
				// If that fails, fall back to getting all text and cleaning it later
				if definitionText == "" {
					definitionText = strings.TrimSpace(li.Text)
				}
				
				// Skip empty definitions
				if definitionText == "" {
					return
				}
				
				// Add definition type prefix if available
				if currentDefinitionType != "" {
					definitionText = "(" + currentDefinitionType + ") " + definitionText
				}
				
				// Add the definition
				newWord.Definitions = append(newWord.Definitions, definitionText)
				currentDefinitionIndex = len(newWord.Definitions) - 1
				w.logger.Debug().Int("index", currentDefinitionIndex).Str("definition", definitionText).Msg("Found definition")
				
				// Extract examples separately
				li.ForEach("span.example", func(_ int, example *colly.HTMLElement) {
					exampleText := strings.TrimSpace(example.Text)
					if exampleText != "" {
						// Clean up the example text
						exampleText = strings.ReplaceAll(exampleText, "« ", "")
						exampleText = strings.ReplaceAll(exampleText, " »", "")
						w.logger.Debug().Str("example", exampleText).Msg("Found example")
						newWord.Examples = append(newWord.Examples, exampleText)
					}
				})
				
				foundDefinitions = true
			})
		}
	})
	
	// Additional backup method to find definitions in paragraphs
	c.OnHTML("div.mw-heading-3:has(span.titredef) ~ p, h3:has(span.titredef) ~ p", func(e *colly.HTMLElement) {
		// Only process if we haven't found definitions yet
		if !foundDefinitions {
			definitionText := strings.TrimSpace(e.Text)
			
			// Skip empty definitions or paragraphs that are likely not definitions
			if definitionText == "" || len(definitionText) < 5 {
				return
			}
			
			w.logger.Debug().Msg("Found definition in paragraph")
			
			// Add definition type prefix if available
			if currentDefinitionType != "" {
				definitionText = "(" + currentDefinitionType + ") " + definitionText
			}
			
			// Add the definition
			newWord.Definitions = append(newWord.Definitions, definitionText)
			w.logger.Debug().Int("index", len(newWord.Definitions)-1).Str("definition", definitionText).Msg("Found definition in paragraph")
			
			foundDefinitions = true
		}
	})

	// Extract synonyms
	c.OnHTML("div.mw-heading-4:has(span.titresyno) + ul, h4:has(span.titresyno) + ul", func(e *colly.HTMLElement) {
		e.ForEach("li", func(_ int, li *colly.HTMLElement) {
			synonym := strings.TrimSpace(li.Text)
			if synonym != "" {
				w.logger.Debug().Str("synonym", synonym).Msg("Found synonym")
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
	
	// Remove duplicate definitions and examples
	newWord.Definitions = removeDuplicates(newWord.Definitions)
	newWord.Examples = removeDuplicates(newWord.Examples)

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

// Helper function to remove duplicate strings from a slice
func removeDuplicates(strSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	
	for _, entry := range strSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
