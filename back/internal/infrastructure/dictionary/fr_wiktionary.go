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

	// Structure to store table of contents information
	type TOCSection struct {
		ID       string
		Title    string
		Level    int
		SubItems []*TOCSection
	}

	// Map to store section IDs for easier lookup
	sectionIDs := make(map[string]string)
	
	// Track if we're in the French section
	var inFrenchSection bool = false
	var frenchSectionID string
	
	// Parse the table of contents to get section IDs and structure
	c.OnHTML("#mw-panel-toc-list", func(e *colly.HTMLElement) {
		w.logger.Debug().Msg("Found table of contents")
		
		// Find the French section in the TOC
		e.ForEach("li", func(_ int, li *colly.HTMLElement) {
			sectionID := li.Attr("id")
			if strings.HasPrefix(sectionID, "toc-") {
				// Extract the language from the section
				langSpan := li.ChildText("span.sectionlangue")
				if langSpan == "Français" {
					w.logger.Debug().Str("sectionID", sectionID).Msg("Found French section in TOC")
					inFrenchSection = true
					frenchSectionID = strings.TrimPrefix(sectionID, "toc-")
					
					// Now parse the subsections of the French section
					li.ForEach("ul li", func(_ int, subLi *colly.HTMLElement) {
						subSectionID := subLi.Attr("id")
						if strings.HasPrefix(subSectionID, "toc-") {
							subSectionTitle := subLi.ChildText("span.vector-toc-text span:last-child")
							actualID := strings.TrimPrefix(subSectionID, "toc-")
							sectionIDs[subSectionTitle] = actualID
							w.logger.Debug().Str("title", subSectionTitle).Str("id", actualID).Msg("Found subsection")
							
							// Parse deeper levels (definitions, synonyms, etc.)
							subLi.ForEach("ul li", func(_ int, subSubLi *colly.HTMLElement) {
								subSubSectionID := subSubLi.Attr("id")
								if strings.HasPrefix(subSubSectionID, "toc-") {
									subSubSectionTitle := subSubLi.ChildText("span.vector-toc-text span:last-child")
									actualSubID := strings.TrimPrefix(subSubSectionID, "toc-")
									sectionIDs[subSubSectionTitle] = actualSubID
									w.logger.Debug().Str("title", subSubSectionTitle).Str("id", actualSubID).Msg("Found sub-subsection")
								}
							})
						}
					})
				}
			}
		})
	})
	
	// Extract word information from the flextable (gender, plural)
	c.OnHTML("table.flextable-fr-mfsp", func(e *colly.HTMLElement) {
		if !inFrenchSection {
			return
		}
		
		w.logger.Debug().Msg("Found flextable with word information")
		
		// Extract plural form
		e.ForEach("tr", func(_ int, tr *colly.HTMLElement) {
			if strings.Contains(tr.Text, "Pluriel") {
				tr.ForEach("td", func(_ int, td *colly.HTMLElement) {
					plural := strings.TrimSpace(td.Text)
					if plural != "" && plural != newWord.Text {
						w.logger.Debug().Str("plural", plural).Msg("Found plural form")
						newWord.Translations["plural"] = plural
					}
				})
			}
		})
	})
	
	// Extract gender information
	c.OnHTML("p span.ligne-de-forme", func(e *colly.HTMLElement) {
		if !inFrenchSection {
			return
		}
		
		wordType := strings.TrimSpace(e.Text)
		if strings.Contains(wordType, "masculin") || strings.Contains(wordType, "féminin") {
			w.logger.Debug().Str("wordType", wordType).Msg("Found word type")
			newWord.Gender = wordType
		}
	})
	
	// Extract pronunciation
	c.OnHTML("span.API", func(e *colly.HTMLElement) {
		if !inFrenchSection {
			return
		}
		
		pronunciation := strings.TrimSpace(e.Text)
		if strings.HasPrefix(pronunciation, "\\") && strings.HasSuffix(pronunciation, "\\") {
			w.logger.Debug().Str("pronunciation", pronunciation).Msg("Found pronunciation")
			newWord.SetPronunciation("ipa", pronunciation)
		}
	})
	
	// Extract etymology
	c.OnHTML("#Étymologie, #Etymology", func(e *colly.HTMLElement) {
		if !inFrenchSection {
			return
		}
		
		// Get the content after the etymology heading
		etymologyText := ""
		nextElem := e.DOM.Parent().Next()
		if nextElem.Is("dl") {
			etymologyText = strings.TrimSpace(nextElem.Text())
		} else if nextElem.Is("p") {
			etymologyText = strings.TrimSpace(nextElem.Text())
		}
		
		if etymologyText != "" {
			w.logger.Debug().Str("etymology", etymologyText).Msg("Found etymology")
			newWord.Etymology = etymologyText
		}
	})
	
	// Extract definitions from the Nom_commun (Common noun) section
	c.OnHTML("#Nom_commun, #Noun", func(e *colly.HTMLElement) {
		if !inFrenchSection {
			return
		}
		
		w.logger.Debug().Msg("Found noun section")
		
		// Look for the ordered list that contains definitions
		var definitionList *colly.HTMLElement
		nextElem := e.DOM.Parent().Next()
		
		// Try to find the definition list (could be directly after the heading or after a paragraph)
		for i := 0; i < 3 && nextElem.Length() > 0; i++ {
			if nextElem.Is("ol") {
				definitionList = &colly.HTMLElement{
					DOM: nextElem,
				}
				break
			} else if nextElem.Is("p") && nextElem.Next().Is("ol") {
				definitionList = &colly.HTMLElement{
					DOM: nextElem.Next(),
				}
				break
			}
			nextElem = nextElem.Next()
		}
		
		if definitionList != nil {
			// Process each list item as a definition
			definitionList.DOM.Find("li").Each(func(i int, liSelection *goquery.Selection) {
				// Create a colly HTMLElement for the list item
				li := &colly.HTMLElement{
					DOM: liSelection,
				}
				
				// Extract definition type if present
				defType := ""
				li.ForEach("span.emploi", func(_ int, span *colly.HTMLElement) {
					defType = strings.TrimSpace(span.Text)
					defType = strings.TrimPrefix(defType, "(")
					defType = strings.TrimSuffix(defType, ")")
				})
				
				// Get the main definition text - need to use goquery for this
				definitionText := strings.TrimSpace(liSelection.Contents().Not("ul").Not("span.example").Text())
				
				// If empty, try getting the full text and cleaning it
				if definitionText == "" {
					fullText := strings.TrimSpace(li.Text)
					
					// Try to separate definition from examples
					if idx := strings.Index(fullText, " : "); idx > 0 {
						definitionText = strings.TrimSpace(fullText[:idx])
					} else if idx := strings.Index(fullText, " — "); idx > 0 {
						definitionText = strings.TrimSpace(fullText[:idx])
					} else if idx := strings.Index(fullText, "« "); idx > 0 {
						definitionText = strings.TrimSpace(fullText[:idx])
					} else {
						definitionText = fullText
					}
				}
				
				// Skip empty definitions
				if definitionText == "" {
					return
				}
				
				// Add definition type prefix if available
				if defType != "" {
					definitionText = "(" + defType + ") " + definitionText
				}
				
				// Add the definition
				newWord.Definitions = append(newWord.Definitions, definitionText)
				w.logger.Debug().Int("index", len(newWord.Definitions)-1).Str("definition", definitionText).Msg("Found definition")
				
				// Extract examples
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
	
	// Extract synonyms
	c.OnHTML("span.titresyno", func(e *colly.HTMLElement) {
		if !inFrenchSection {
			return
		}
		
		w.logger.Debug().Msg("Found synonyms section")
		
		// Look for the unordered list that contains synonyms
		synonymList := e.DOM.Parent().Parent().Next()
		if synonymList.Is("ul") {
			// Create a colly HTMLElement for the synonym list
			synonymListElem := &colly.HTMLElement{
				DOM: synonymList,
			}
			
			synonymListElem.ForEach("li", func(_ int, li *colly.HTMLElement) {
				synonym := strings.TrimSpace(li.Text)
				if synonym != "" {
					w.logger.Debug().Str("synonym", synonym).Msg("Found synonym")
					newWord.Synonyms = append(newWord.Synonyms, synonym)
				}
			})
		}
	})
	
	// Extract antonyms
	c.OnHTML("span.titreanto", func(e *colly.HTMLElement) {
		if !inFrenchSection {
			return
		}
		
		w.logger.Debug().Msg("Found antonyms section")
		
		// Look for the unordered list that contains antonyms
		antonymList := e.DOM.Parent().Parent().Next()
		if antonymList.Is("ul") {
			// Create a colly HTMLElement for the antonym list
			antonymListElem := &colly.HTMLElement{
				DOM: antonymList,
			}
			
			antonymListElem.ForEach("li", func(_ int, li *colly.HTMLElement) {
				antonym := strings.TrimSpace(li.Text)
				if antonym != "" {
					w.logger.Debug().Str("antonym", antonym).Msg("Found antonym")
					newWord.Antonyms = append(newWord.Antonyms, antonym)
				}
			})
		}
	})
	
	// Extract translations
	c.OnHTML("span.titretrad", func(e *colly.HTMLElement) {
		if !inFrenchSection {
			return
		}
		
		w.logger.Debug().Msg("Found translations section")
		
		// Translations are in a complex structure, navigate to find them
		translationsDiv := e.DOM.Parent().Parent().Next()
		if translationsDiv.HasClass("boite") {
			// Create a colly HTMLElement for the translations div
			translationsDivElem := &colly.HTMLElement{
				DOM: translationsDiv,
			}
			
			translationsDivElem.ForEach("li", func(_ int, li *colly.HTMLElement) {
				// We need to use goquery for some operations
				liSelection := li.DOM
				
				langSpan := liSelection.Find("span[class^='trad-']").First()
				langName := strings.TrimSpace(langSpan.Text())
				
				// Extract the translation text - need a different approach
				translationText := ""
				
				// Get the full text and try to extract the translation
				fullText := strings.TrimSpace(li.Text)
				
				// Remove the language name and colon
				for _, prefix := range []string{"Allemand", "Anglais", "Espagnol", "Italien", "Portugais", "Roumain"} {
					if strings.HasPrefix(fullText, prefix) {
						fullText = strings.TrimSpace(strings.TrimPrefix(fullText, prefix))
						break
					}
				}
				
				// Remove the colon
				fullText = strings.TrimSpace(strings.TrimPrefix(fullText, ":"))
				
				if fullText != "" {
					translationText = fullText
				}
				
				// Clean up the translation text
				translationText = strings.TrimSpace(translationText)
				translationText = strings.TrimPrefix(translationText, ":")
				translationText = strings.TrimSpace(translationText)
				
				if langName != "" && translationText != "" {
					langCode := ""
					switch {
					case strings.Contains(langName, "Allemand"):
						langCode = "de"
					case strings.Contains(langName, "Anglais"):
						langCode = "en"
					case strings.Contains(langName, "Espagnol"):
						langCode = "es"
					case strings.Contains(langName, "Italien"):
						langCode = "it"
					case strings.Contains(langName, "Portugais"):
						langCode = "pt"
					case strings.Contains(langName, "Roumain"):
						langCode = "ro"
					}
					
					if langCode != "" {
						w.logger.Debug().Str("language", langCode).Str("translation", translationText).Msg("Found translation")
						newWord.Translations[langCode] = translationText
					}
				}
			})
		}
	})
	
	// Fallback method for definitions if none were found
	c.OnHTML("body", func(e *colly.HTMLElement) {
		// Only run this if we haven't found definitions yet
		if !foundDefinitions && inFrenchSection {
			w.logger.Warn().Msg("No definitions found with primary selectors, trying fallback")
			
			// Try to find any ordered list in the French section
			e.ForEach("ol li", func(_ int, li *colly.HTMLElement) {
				if !foundDefinitions {
					definitionText := strings.TrimSpace(li.Text)
					if definitionText != "" && len(definitionText) > 10 {
						w.logger.Debug().Str("definition", definitionText).Msg("Found definition with fallback method")
						newWord.Definitions = append(newWord.Definitions, definitionText)
						foundDefinitions = true
					}
				}
			})
			
			// If still no definitions, try paragraphs
			if !foundDefinitions {
				e.ForEach("p", func(_ int, p *colly.HTMLElement) {
					if !foundDefinitions {
						fullText := strings.TrimSpace(p.Text)
						if fullText != "" && len(fullText) > 10 && !strings.HasPrefix(fullText, "From") {
							w.logger.Debug().Str("definition", fullText).Msg("Found definition in paragraph with fallback method")
							newWord.Definitions = append(newWord.Definitions, fullText)
							foundDefinitions = true
						}
					}
				})
			}
		}
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

	// If still no definitions, return error
	if len(newWord.Definitions) == 0 {
		w.logger.Warn().Str("text", text).Str("language", language).Msg("No word data found")
		return nil, fmt.Errorf("no word data found: %w", word.ErrWordNotFound)
	}

	// Remove duplicate definitions and examples
	newWord.Definitions = removeDuplicates(newWord.Definitions)
	newWord.Examples = removeDuplicates(newWord.Examples)
	newWord.Synonyms = removeDuplicates(newWord.Synonyms)
	newWord.Antonyms = removeDuplicates(newWord.Antonyms)

	w.logger.Debug().
		Str("text", text).
		Str("language", language).
		Int("definitions", len(newWord.Definitions)).
		Int("examples", len(newWord.Examples)).
		Int("synonyms", len(newWord.Synonyms)).
		Int("antonyms", len(newWord.Antonyms)).
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

// Helper function to check if a string is in a slice
func containsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}
