package dictionary

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/rs/zerolog"

	wordDomain "voconsteroid/internal/domain/word"
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

// PageStructure represents the structure of a Wiktionary page
type PageStructure struct {
	// Map of section titles to their IDs
	SectionIDs map[string]string
	// Flag indicating if the French section was found
	HasFrenchSection bool
	// Map of section types to their IDs (noun, verb, etc.)
	WordTypeSections map[string]string
	// Map of other section types (synonyms, antonyms, etc.)
	OtherSections map[string]string
}

// FetchRelatedWords retrieves words related to the given word from French Wiktionary
func (w *FrenchWiktionaryAPI) FetchRelatedWords(ctx context.Context, word *wordDomain.Word) (*wordDomain.RelatedWords, error) {
	w.logger.Debug().Str("word", word.Text).Str("language", word.Language).Msg("Fetching related words from French Wiktionary")

	// Create a new RelatedWords object with the source word
	relatedWords := &wordDomain.RelatedWords{
		SourceWord: word,
		Synonyms:   []*wordDomain.Word{},
		Antonyms:   []*wordDomain.Word{},
	}

	// Track which words we've already processed to avoid duplicates
	processedSynonyms := make(map[string]bool)
	processedAntonyms := make(map[string]bool)

	// If the word already has synonyms or antonyms, fetch them
	for _, synonym := range word.Synonyms {
		if synonym == "" || processedSynonyms[synonym] {
			continue
		}
		processedSynonyms[synonym] = true

		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Fetch the synonym word
			synonymWord, err := w.FetchWord(ctx, synonym, word.Language)
			if err == nil && synonymWord != nil {
				relatedWords.Synonyms = append(relatedWords.Synonyms, synonymWord)
			} else {
				// If we can't fetch the full word, create a minimal one
				minimalWord := wordDomain.NewWord(synonym, word.Language)
				relatedWords.Synonyms = append(relatedWords.Synonyms, minimalWord)
			}
		}
	}

	// Do the same for antonyms
	for _, antonym := range word.Antonyms {
		if antonym == "" || processedAntonyms[antonym] {
			continue
		}
		processedAntonyms[antonym] = true

		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Fetch the antonym word
			antonymWord, err := w.FetchWord(ctx, antonym, word.Language)
			if err == nil && antonymWord != nil {
				relatedWords.Antonyms = append(relatedWords.Antonyms, antonymWord)
			} else {
				// If we can't fetch the full word, create a minimal one
				minimalWord := wordDomain.NewWord(antonym, word.Language)
				relatedWords.Antonyms = append(relatedWords.Antonyms, minimalWord)
			}
		}
	}

	w.logger.Debug().
		Str("word", word.Text).
		Int("synonyms", len(relatedWords.Synonyms)).
		Int("antonyms", len(relatedWords.Antonyms)).
		Msg("Successfully fetched related words")

	return relatedWords, nil
}

// FetchWord retrieves word information from French Wiktionary by scraping the web page
func (w *FrenchWiktionaryAPI) FetchWord(ctx context.Context, text, language string) (*wordDomain.Word, error) {
	w.logger.Debug().Str("text", text).Str("language", language).Msg("Fetching word from French Wiktionary")

	// Create a new word
	newWord := wordDomain.NewWord(text, language)

	// First pass: Extract the table of contents to understand the page structure
	pageStructure, err := w.extractPageStructure(ctx, text, language)
	if err != nil {
		return nil, err
	}

	// If no French section was found, return an error
	if !pageStructure.HasFrenchSection {
		w.logger.Warn().Str("text", text).Str("language", language).Msg("No French section found")
		return nil, fmt.Errorf("no French section found: %w", wordDomain.ErrWordNotFound)
	}

	w.logger.Debug().
		Str("text", text).
		Int("sections", len(pageStructure.SectionIDs)).
		Int("wordTypeSections", len(pageStructure.WordTypeSections)).
		Int("otherSections", len(pageStructure.OtherSections)).
		Msg("Page structure extracted")

	// Second pass: Extract the word information based on the page structure
	foundDefinitions, err := w.extractWordInformation(ctx, newWord, pageStructure, text, language)
	if err != nil {
		return nil, err
	}

	// If still no definitions, return error
	if !foundDefinitions || len(newWord.Definitions) == 0 {
		w.logger.Warn().Str("text", text).Str("language", language).Msg("No word data found")
		return nil, fmt.Errorf("no word data found: %w", wordDomain.ErrWordNotFound)
	}

	w.logger.Debug().
		Str("text", text).
		Str("language", language).
		Int("definitions", len(newWord.Definitions)).
		Int("examples", len(newWord.Examples)).
		Int("synonyms", len(newWord.Synonyms)).
		Int("antonyms", len(newWord.Antonyms)).
		Str("etymology", newWord.Etymology).
		Interface("pronunciation", newWord.Pronunciation).
		Msg("Successfully fetched word data from French Wiktionary")

	return newWord, nil
}

// extractPageStructure extracts the structure of the page by analyzing the table of contents
func (w *FrenchWiktionaryAPI) extractPageStructure(ctx context.Context, text, language string) (*PageStructure, error) {
	// Create a new collector for TOC extraction
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		colly.MaxDepth(1),
	)

	// Set timeout
	c.SetRequestTimeout(10 * time.Second)

	// Add rate limiting to avoid being blocked
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*.wiktionary.org",
		Delay:       1 * time.Second,
		RandomDelay: 500 * time.Millisecond,
	})

	// Initialize the page structure
	pageStructure := &PageStructure{
		SectionIDs:       make(map[string]string),
		HasFrenchSection: false,
		WordTypeSections: make(map[string]string),
		OtherSections:    make(map[string]string),
	}

	// Setup the callback for the table of contents
	c.OnHTML("#mw-panel-toc-list", func(e *colly.HTMLElement) {
		w.logger.Debug().Msg("Found table of contents")

		// Find the French section in the TOC
		e.ForEach("li", func(_ int, li *colly.HTMLElement) {
			sectionID := li.Attr("id")
			sectionTitle := li.ChildText("span.vector-toc-text span:last-child")

			if strings.HasPrefix(sectionID, "toc-") {
				actualID := strings.TrimPrefix(sectionID, "toc-")
				pageStructure.SectionIDs[sectionTitle] = actualID

				// Check if this is the French section
				if strings.Contains(sectionTitle, "Français") {
					w.logger.Debug().Str("sectionID", sectionID).Msg("Found French section in TOC")
					pageStructure.HasFrenchSection = true

					// Now parse the subsections of the French section
					li.ForEach("ul li", func(_ int, subLi *colly.HTMLElement) {
						subSectionID := subLi.Attr("id")
						subSectionTitle := subLi.ChildText("span.vector-toc-text span:last-child")

						if strings.HasPrefix(subSectionID, "toc-") {
							actualSubID := strings.TrimPrefix(subSectionID, "toc-")
							pageStructure.SectionIDs[subSectionTitle] = actualSubID

							// Categorize the section
							switch {
							case strings.Contains(subSectionTitle, "Nom") ||
								strings.Contains(subSectionTitle, "Verbe") ||
								strings.Contains(subSectionTitle, "Adjectif") ||
								strings.Contains(subSectionTitle, "Adverbe"):
								pageStructure.WordTypeSections[subSectionTitle] = actualSubID
							case strings.Contains(subSectionTitle, "Étymologie"):
								pageStructure.OtherSections["etymology"] = actualSubID
							case strings.Contains(subSectionTitle, "Prononciation"):
								pageStructure.OtherSections["pronunciation"] = actualSubID
							default:
								// Other sections
								pageStructure.OtherSections[subSectionTitle] = actualSubID
							}

							w.logger.Debug().Str("title", subSectionTitle).Str("id", actualSubID).Msg("Found subsection")

							// Parse deeper levels (definitions, synonyms, etc.)
							subLi.ForEach("ul li", func(_ int, subSubLi *colly.HTMLElement) {
								subSubSectionID := subSubLi.Attr("id")
								subSubSectionTitle := subSubLi.ChildText("span.vector-toc-text span:last-child")

								if strings.HasPrefix(subSubSectionID, "toc-") {
									actualSubSubID := strings.TrimPrefix(subSubSectionID, "toc-")
									pageStructure.SectionIDs[subSubSectionTitle] = actualSubSubID

									// Categorize the subsection
									switch {
									case strings.Contains(subSubSectionTitle, "Synonymes"):
										pageStructure.OtherSections["synonyms"] = actualSubSubID
									case strings.Contains(subSubSectionTitle, "Antonymes"):
										pageStructure.OtherSections["antonyms"] = actualSubSubID
									case strings.Contains(subSubSectionTitle, "Traductions"):
										pageStructure.OtherSections["translations"] = actualSubSubID
									case strings.Contains(subSubSectionTitle, "Dérivés"):
										pageStructure.OtherSections["derivatives"] = actualSubSubID
									case strings.Contains(subSubSectionTitle, "Apparentés"):
										pageStructure.OtherSections["related"] = actualSubSubID
									case strings.Contains(subSubSectionTitle, "Variantes"):
										pageStructure.OtherSections["variants"] = actualSubSubID
									case strings.Contains(subSubSectionTitle, "Voir aussi"):
										pageStructure.OtherSections["see_also"] = actualSubSubID
									case strings.Contains(subSubSectionTitle, "Références"):
										pageStructure.OtherSections["references"] = actualSubSubID
									}

									w.logger.Debug().Str("title", subSubSectionTitle).Str("id", actualSubSubID).Msg("Found sub-subsection")
								}
							})
						}
					})
				}
			}
		})
	})

	// If no TOC is found, we'll try to find sections directly
	c.OnHTML("body", func(e *colly.HTMLElement) {
		// Only run this if we haven't found a French section yet
		if !pageStructure.HasFrenchSection {
			w.logger.Debug().Msg("No TOC found, looking for sections directly")

			// Look for section headings
			e.ForEach("h2, h3, h4", func(_ int, heading *colly.HTMLElement) {
				headingText := strings.TrimSpace(heading.Text)
				headingID := heading.Attr("id")

				// Check if this is the French section
				if strings.Contains(headingText, "Français") {
					w.logger.Debug().Str("headingID", headingID).Msg("Found French section directly")
					pageStructure.HasFrenchSection = true
					pageStructure.SectionIDs[headingText] = headingID
				}

				// If we're already in the French section, categorize subsections
				if pageStructure.HasFrenchSection && heading.Name == "h3" {
					switch {
					case strings.Contains(headingText, "Nom") ||
						strings.Contains(headingText, "Verbe") ||
						strings.Contains(headingText, "Adjectif") ||
						strings.Contains(headingText, "Adverbe"):
						pageStructure.WordTypeSections[headingText] = headingID
					case strings.Contains(headingText, "Étymologie"):
						pageStructure.OtherSections["etymology"] = headingID
					case strings.Contains(headingText, "Prononciation"):
						pageStructure.OtherSections["pronunciation"] = headingID
					default:
						// Other sections
						pageStructure.OtherSections[headingText] = headingID
					}

					w.logger.Debug().Str("title", headingText).Str("id", headingID).Msg("Found section directly")
				}

				// If we're in a word type section, look for subsections
				if pageStructure.HasFrenchSection && heading.Name == "h4" {
					switch {
					case strings.Contains(headingText, "Synonymes"):
						pageStructure.OtherSections["synonyms"] = headingID
					case strings.Contains(headingText, "Antonymes"):
						pageStructure.OtherSections["antonyms"] = headingID
					case strings.Contains(headingText, "Traductions"):
						pageStructure.OtherSections["translations"] = headingID
					}

					w.logger.Debug().Str("title", headingText).Str("id", headingID).Msg("Found subsection directly")
				}
			})
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
			w.logger.Error().Err(err).Str("url", url).Msg("Failed to visit page for TOC extraction")
			return nil, fmt.Errorf("failed to visit page: %w", err)
		}
	}

	// Wait until scraping is finished
	c.Wait()

	return pageStructure, nil
}

// extractWordInformation extracts the word information based on the page structure
func (w *FrenchWiktionaryAPI) extractWordInformation(ctx context.Context, word *wordDomain.Word, pageStructure *PageStructure, text, language string) (bool, error) {
	// Create a new collector for content extraction
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		colly.MaxDepth(1),
	)

	// Set timeout
	c.SetRequestTimeout(10 * time.Second)

	// Add rate limiting to avoid being blocked
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*.wiktionary.org",
		Delay:       1 * time.Second,
		RandomDelay: 500 * time.Millisecond,
	})

	// Track if we found any definitions
	foundDefinitions := false

	// Track seen examples to avoid duplicates
	seenExamples := make(map[string]bool)

	// Setup callbacks for basic word information
	w.setupWordFormsCallback(c, word)
	w.setupGenderCallback(c, word)
	w.setupPronunciationCallback(c, word, pageStructure)

	// Setup callbacks for sections based on the page structure
	if etymologyID, ok := pageStructure.OtherSections["etymology"]; ok {
		w.setupEtymologyCallback(c, word, etymologyID)
	}

	// Setup callbacks for word type sections (noun, verb, etc.)
	for sectionTitle, sectionID := range pageStructure.WordTypeSections {
		wordType := w.determineWordType(sectionTitle)
		w.setupDefinitionsCallback(c, word, sectionID, wordType, &foundDefinitions, seenExamples)
	}

	// Setup callbacks for other sections
	if synonymsID, ok := pageStructure.OtherSections["synonyms"]; ok {
		w.setupSynonymsCallback(c, word, synonymsID)
	}

	if antonymsID, ok := pageStructure.OtherSections["antonyms"]; ok {
		w.setupAntonymsCallback(c, word, antonymsID)
	}

	if translationsID, ok := pageStructure.OtherSections["translations"]; ok {
		w.setupTranslationsCallback(c, word, translationsID)
	}

	// Setup fallback callback for definitions if none are found
	w.setupFallbackDefinitionsCallback(c, word, &foundDefinitions, seenExamples)

	// Build URL for the web page
	baseURL := w.getBaseURL(language)
	url := fmt.Sprintf("%s/%s", baseURL, text)

	// Check if context is done
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		// Visit the page
		err := c.Visit(url)
		if err != nil {
			w.logger.Error().Err(err).Str("url", url).Msg("Failed to visit page for word information extraction")
			return false, fmt.Errorf("failed to visit page: %w", err)
		}
	}

	// Wait until scraping is finished
	c.Wait()

	return foundDefinitions, nil
}

// determineWordType determines the word type from a section title
func (w *FrenchWiktionaryAPI) determineWordType(sectionTitle string) string {
	switch {
	case strings.Contains(sectionTitle, "Nom"):
		return "noun"
	case strings.Contains(sectionTitle, "Verbe"):
		return "verb"
	case strings.Contains(sectionTitle, "Adjectif"):
		return "adjective"
	case strings.Contains(sectionTitle, "Adverbe"):
		return "adverb"
	case strings.Contains(sectionTitle, "Pronom"):
		return "pronoun"
	case strings.Contains(sectionTitle, "Préposition"):
		return "preposition"
	case strings.Contains(sectionTitle, "Conjonction"):
		return "conjunction"
	case strings.Contains(sectionTitle, "Interjection"):
		return "interjection"
	default:
		return ""
	}
}

// setupWordFormsCallback sets up the callback for extracting word forms
func (w *FrenchWiktionaryAPI) setupWordFormsCallback(c *colly.Collector, word *wordDomain.Word) {
	c.OnHTML("table.flextable-fr-mfsp", func(e *colly.HTMLElement) {
		w.logger.Debug().Msg("Found flextable with word information")

		// Extract plural form
		e.ForEach("tr", func(_ int, tr *colly.HTMLElement) {
			if strings.Contains(tr.Text, "Pluriel") {
				tr.ForEach("td", func(_ int, td *colly.HTMLElement) {
					plural := strings.TrimSpace(td.Text)
					if plural != "" && plural != word.Text {
						w.logger.Debug().Str("plural", plural).Msg("Found plural form")
						word.Translations["plural"] = plural

						// Add this as a word form
						pluralAttributes := map[string]string{"number": "plural"}
						word.AddWordForm(plural, pluralAttributes, false)
					}
				})
			}
		})
	})
}

// setupGenderCallback sets up the callback for extracting gender information
func (w *FrenchWiktionaryAPI) setupGenderCallback(c *colly.Collector, word *wordDomain.Word) {
	c.OnHTML("p span.ligne-de-forme", func(e *colly.HTMLElement) {
		wordType := strings.TrimSpace(e.Text)
		if strings.Contains(wordType, "masculin") || strings.Contains(wordType, "féminin") {
			w.logger.Debug().Str("wordType", wordType).Msg("Found word type")
			word.Gender = wordType

			// If this is a form of another word, try to extract the lemma
			if strings.Contains(wordType, "de ") {
				parts := strings.Split(wordType, "de ")
				if len(parts) > 1 {
					lemma := strings.TrimSpace(parts[1])
					if lemma != "" && lemma != word.Text {
						w.logger.Debug().Str("lemma", lemma).Msg("Found lemma")
						word.SetLemma(lemma)
					}
				}
			}
		}
	})
}

// setupPronunciationCallback sets up the callback for extracting pronunciation
func (w *FrenchWiktionaryAPI) setupPronunciationCallback(c *colly.Collector, word *wordDomain.Word, pageStructure *PageStructure) {
	// Try both the section ID and the general API span
	if pronunciationID, ok := pageStructure.OtherSections["pronunciation"]; ok {
		selector := fmt.Sprintf("#%s", pronunciationID)
		c.OnHTML(selector, func(e *colly.HTMLElement) {
			w.logger.Debug().Str("selector", selector).Msg("Found pronunciation section")

			// Look for IPA notation in this section
			e.DOM.Parent().Next().Find("span.API").Each(func(_ int, span *goquery.Selection) {
				pronunciation := strings.TrimSpace(span.Text())
				if strings.HasPrefix(pronunciation, "\\") && strings.HasSuffix(pronunciation, "\\") {
					w.logger.Debug().Str("pronunciation", pronunciation).Msg("Found pronunciation in section")
					if word.Pronunciation == nil {
						word.Pronunciation = make(map[string]string)
					}
					word.Pronunciation["ipa"] = pronunciation
				}
			})
		})
	}

	// Also try the general API span as a fallback
	c.OnHTML("span.API", func(e *colly.HTMLElement) {
		// Only set if we haven't already found it in a section
		if word.Pronunciation == nil || word.Pronunciation["ipa"] == "" {
			pronunciation := strings.TrimSpace(e.Text)
			if strings.HasPrefix(pronunciation, "\\") && strings.HasSuffix(pronunciation, "\\") {
				w.logger.Debug().Str("pronunciation", pronunciation).Msg("Found pronunciation")
				if word.Pronunciation == nil {
					word.Pronunciation = make(map[string]string)
				}
				word.Pronunciation["ipa"] = pronunciation
			}
		}
	})
}

// setupEtymologyCallback sets up the callback for extracting etymology
func (w *FrenchWiktionaryAPI) setupEtymologyCallback(c *colly.Collector, word *wordDomain.Word, etymologyID string) {
	selector := fmt.Sprintf("#%s", etymologyID)
	c.OnHTML(selector, func(e *colly.HTMLElement) {
		w.logger.Debug().Str("selector", selector).Msg("Found etymology section")

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
			word.Etymology = etymologyText
		}
	})
}

// setupDefinitionsCallback sets up the callback for extracting definitions
func (w *FrenchWiktionaryAPI) setupDefinitionsCallback(c *colly.Collector, word *wordDomain.Word, sectionID, wordType string, foundDefinitions *bool, seenExamples map[string]bool) {
	selector := fmt.Sprintf("#%s", sectionID)
	c.OnHTML(selector, func(e *colly.HTMLElement) {
		w.logger.Debug().Str("selector", selector).Str("wordType", wordType).Msg("Found word type section")

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

				// Collect examples for this definition
				examples := []string{}
				liSelection.Find("span.example").Each(func(_ int, exampleSpan *goquery.Selection) {
					exampleText := strings.TrimSpace(exampleSpan.Text())
					if exampleText != "" {
						// Clean up the example text
						exampleText = strings.ReplaceAll(exampleText, "« ", "")
						exampleText = strings.ReplaceAll(exampleText, " »", "")

						// Only add if we haven't seen this example before
						if !seenExamples[exampleText] {
							w.logger.Debug().Str("example", exampleText).Msg("Found example")
							examples = append(examples, exampleText)
							seenExamples[exampleText] = true

							// Also add to the word's general examples
							word.Examples = append(word.Examples, exampleText)
						}
					}
				})

				// Add the definition using the entity method
				word.Definitions = append(word.Definitions, wordDomain.Definition{
					Text:     definitionText,
					WordType: wordType,
					Examples: examples,
				})
				w.logger.Debug().Int("index", len(word.Definitions)-1).Str("definition", definitionText).Msg("Found definition")

				*foundDefinitions = true
			})
		}
	})
}

// setupSynonymsCallback sets up the callback for extracting synonyms
func (w *FrenchWiktionaryAPI) setupSynonymsCallback(c *colly.Collector, word *wordDomain.Word, synonymsID string) {
	// Use the section ID if available
	selector := fmt.Sprintf("#%s", synonymsID)
	c.OnHTML(selector, func(e *colly.HTMLElement) {
		w.logger.Debug().Str("selector", selector).Msg("Found synonyms section")

		// Look for the unordered list that contains synonyms
		var synonymList *goquery.Selection

		// Look for the next ul
		synonymList = e.DOM.Parent().Next()
		if !synonymList.Is("ul") {
			// Try the next element
			synonymList = synonymList.Next()
			if !synonymList.Is("ul") {
				return
			}
		}

		// Create a colly HTMLElement for the synonym list
		synonymListElem := &colly.HTMLElement{
			DOM: synonymList,
		}

		synonymListElem.ForEach("li", func(_ int, li *colly.HTMLElement) {
			synonym := strings.TrimSpace(li.Text)
			if synonym != "" {
				w.logger.Debug().Str("synonym", synonym).Msg("Found synonym")
				word.AddSynonym(synonym)
			}
		})
	})

	// Also try the span.titresyno as a fallback
	c.OnHTML("span.titresyno", func(e *colly.HTMLElement) {
		// Only process if we don't already have synonyms
		if len(word.Synonyms) == 0 {
			w.logger.Debug().Msg("Found synonyms section via span.titresyno")

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
						word.AddSynonym(synonym)
					}
				})
			}
		}
	})
}

// setupAntonymsCallback sets up the callback for extracting antonyms
func (w *FrenchWiktionaryAPI) setupAntonymsCallback(c *colly.Collector, word *wordDomain.Word, antonymsID string) {
	// Use the section ID if available
	selector := fmt.Sprintf("#%s", antonymsID)
	c.OnHTML(selector, func(e *colly.HTMLElement) {
		w.logger.Debug().Str("selector", selector).Msg("Found antonyms section")

		// Look for the unordered list that contains antonyms
		var antonymList *goquery.Selection

		// Look for the next ul
		antonymList = e.DOM.Parent().Next()
		if !antonymList.Is("ul") {
			// Try the next element
			antonymList = antonymList.Next()
			if !antonymList.Is("ul") {
				return
			}
		}

		// Create a colly HTMLElement for the antonym list
		antonymListElem := &colly.HTMLElement{
			DOM: antonymList,
		}

		antonymListElem.ForEach("li", func(_ int, li *colly.HTMLElement) {
			antonym := strings.TrimSpace(li.Text)
			if antonym != "" {
				w.logger.Debug().Str("antonym", antonym).Msg("Found antonym")
				word.AddAntonym(antonym)
			}
		})
	})

	// Also try the span.titreanto as a fallback
	c.OnHTML("span.titreanto", func(e *colly.HTMLElement) {
		// Only process if we don't already have antonyms
		if len(word.Antonyms) == 0 {
			w.logger.Debug().Msg("Found antonyms section via span.titreanto")

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
						word.AddAntonym(antonym)
					}
				})
			}
		}
	})
}

// setupTranslationsCallback sets up the callback for extracting translations
func (w *FrenchWiktionaryAPI) setupTranslationsCallback(c *colly.Collector, word *wordDomain.Word, translationsID string) {
	// Use the section ID if available
	selector := fmt.Sprintf("#%s", translationsID)
	c.OnHTML(selector, func(e *colly.HTMLElement) {
		w.logger.Debug().Str("selector", selector).Msg("Found translations section")

		// Translations are in a complex structure, navigate to find them
		var translationsDiv *goquery.Selection

		// Look for the next div with class boite
		translationsDiv = e.DOM.Parent().Next()
		for i := 0; i < 3 && translationsDiv.Length() > 0; i++ {
			if translationsDiv.HasClass("boite") {
				break
			}
			translationsDiv = translationsDiv.Next()
		}

		if !translationsDiv.HasClass("boite") {
			return
		}

		// Create a colly HTMLElement for the translations div
		translationsDivElem := &colly.HTMLElement{
			DOM: translationsDiv,
		}

		// Track which languages we've already processed
		processedLangs := make(map[string]bool)

		translationsDivElem.ForEach("li", func(_ int, li *colly.HTMLElement) {
			// We need to use goquery for some operations
			liSelection := li.DOM

			langSpan := liSelection.Find("span[class^='trad-']").First()
			langName := strings.TrimSpace(langSpan.Text())

			// Skip if we've already processed this language
			langCode := w.mapLanguageNameToCode(langName)
			if langCode == "" || processedLangs[langCode] {
				return
			}
			processedLangs[langCode] = true

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
				w.logger.Debug().Str("language", langCode).Str("translation", translationText).Msg("Found translation")
				word.Translations[langCode] = translationText
			}
		})
	})

	// Also try the span.titretrad as a fallback
	c.OnHTML("span.titretrad", func(e *colly.HTMLElement) {
		// Only process if we don't already have translations
		if len(word.Translations) <= 1 { // Account for the "plural" translation that might be present
			w.logger.Debug().Msg("Found translations section via span.titretrad")

			// Translations are in a complex structure, navigate to find them
			translationsDiv := e.DOM.Parent().Parent().Next()
			if translationsDiv.HasClass("boite") {
				// Create a colly HTMLElement for the translations div
				translationsDivElem := &colly.HTMLElement{
					DOM: translationsDiv,
				}

				// Track which languages we've already processed
				processedLangs := make(map[string]bool)

				translationsDivElem.ForEach("li", func(_ int, li *colly.HTMLElement) {
					// We need to use goquery for some operations
					liSelection := li.DOM

					langSpan := liSelection.Find("span[class^='trad-']").First()
					langName := strings.TrimSpace(langSpan.Text())

					// Skip if we've already processed this language
					langCode := w.mapLanguageNameToCode(langName)
					if langCode == "" || processedLangs[langCode] {
						return
					}
					processedLangs[langCode] = true

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
						w.logger.Debug().Str("language", langCode).Str("translation", translationText).Msg("Found translation")
						word.Translations[langCode] = translationText
					}
				})
			}
		}
	})
}

// mapLanguageNameToCode maps a language name to its ISO code
func (w *FrenchWiktionaryAPI) mapLanguageNameToCode(langName string) string {
	switch {
	case strings.Contains(langName, "Allemand"):
		return "de"
	case strings.Contains(langName, "Anglais"):
		return "en"
	case strings.Contains(langName, "Espagnol"):
		return "es"
	case strings.Contains(langName, "Italien"):
		return "it"
	case strings.Contains(langName, "Portugais"):
		return "pt"
	case strings.Contains(langName, "Roumain"):
		return "ro"
	default:
		return ""
	}
}

// setupFallbackDefinitionsCallback sets up the callback for fallback definition extraction
func (w *FrenchWiktionaryAPI) setupFallbackDefinitionsCallback(c *colly.Collector, word *wordDomain.Word, foundDefinitions *bool, seenExamples map[string]bool) {
	c.OnHTML("body", func(e *colly.HTMLElement) {
		// Only run this if we haven't found definitions yet
		if !*foundDefinitions {
			w.logger.Warn().Msg("No definitions found with primary selectors, trying fallback")

			// Try to find any ordered list in the French section
			e.ForEach("ol li", func(_ int, li *colly.HTMLElement) {
				if !*foundDefinitions {
					definitionText := strings.TrimSpace(li.Text)
					if definitionText != "" && len(definitionText) > 10 {
						w.logger.Debug().Str("definition", definitionText).Msg("Found definition with fallback method")
						word.Definitions = append(word.Definitions, wordDomain.Definition{
							Text:     definitionText,
							WordType: "",
							Examples: []string{},
						})
						*foundDefinitions = true
					}
				}
			})

			// If still no definitions, try paragraphs
			if !*foundDefinitions {
				e.ForEach("p", func(_ int, p *colly.HTMLElement) {
					if !*foundDefinitions {
						fullText := strings.TrimSpace(p.Text)
						if fullText != "" && len(fullText) > 10 && !strings.HasPrefix(fullText, "From") {
							w.logger.Debug().Str("definition", fullText).Msg("Found definition in paragraph with fallback method")
							word.Definitions = append(word.Definitions, wordDomain.Definition{
								Text:     fullText,
								WordType: "",
								Examples: []string{},
							})
							*foundDefinitions = true
						}
					}
				})
			}
		}
	})
}
