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

// FetchRelatedWords retrieves words related to the given word from French Wiktionary
func (w *FrenchWiktionaryAPI) FetchRelatedWords(ctx context.Context, word *wordDomain.Word) (*wordDomain.RelatedWords, error) {
	w.logger.Debug().Str("word", word.Text).Str("language", word.Language).Msg("Fetching related words from French Wiktionary")

	// Create a new RelatedWords object with the source word
	relatedWords := &wordDomain.RelatedWords{
		SourceWord: word,
		Synonyms:   []*wordDomain.Word{},
		Antonyms:   []*wordDomain.Word{},
	}

	// If the word already has synonyms or antonyms, fetch them
	for _, synonym := range word.Synonyms {
		if synonym == "" {
			continue
		}

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
		if antonym == "" {
			continue
		}

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

	// Create a new collector
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		colly.MaxDepth(1),
	)

	// Set timeout
	c.SetRequestTimeout(10 * time.Second)

	// Create a new word
	newWord := wordDomain.NewWord(text, language)

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

	// Parse the table of contents to get section IDs and structure
	c.OnHTML("#mw-panel-toc-list", func(e *colly.HTMLElement) {
		w.logger.Debug().Msg("Found table of contents")

		// Find the French section in the TOC
		e.ForEach("li", func(_ int, li *colly.HTMLElement) {
			sectionID := li.Attr("id")
			if strings.HasPrefix(sectionID, "toc-") {
				if strings.HasSuffix(sectionID, "Français") {
					w.logger.Debug().Str("sectionID", sectionID).Msg("Found French section in TOC")
					inFrenchSection = true

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

						// Add this as a word form
						pluralAttributes := map[string]string{"number": "plural"}
						newWord.AddWordForm(plural, pluralAttributes, false)
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

			// If this is a form of another word, try to extract the lemma
			if strings.Contains(wordType, "de ") {
				parts := strings.Split(wordType, "de ")
				if len(parts) > 1 {
					lemma := strings.TrimSpace(parts[1])
					if lemma != "" && lemma != newWord.Text {
						w.logger.Debug().Str("lemma", lemma).Msg("Found lemma")
						newWord.SetLemma(lemma)
					}
				}
			}
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
			if newWord.Pronunciation == nil {
				newWord.Pronunciation = make(map[string]string)
			}
			newWord.Pronunciation["ipa"] = pronunciation
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

	// Extract usage notes
	c.OnHTML("#Usages, #Usage_notes", func(e *colly.HTMLElement) {
		if !inFrenchSection {
			return
		}

		w.logger.Debug().Msg("Found usage notes section")

		// Look for the content after the usage notes heading
		nextElem := e.DOM.Parent().Next()

		// Try different possible formats for usage notes
		if nextElem.Is("ul") {
			nextElem.Find("li").Each(func(_ int, li *goquery.Selection) {
				note := strings.TrimSpace(li.Text())
				if note != "" {
					w.logger.Debug().Str("usageNote", note).Msg("Found usage note")
					newWord.AddUsageNote(note)
				}
			})
		} else if nextElem.Is("p") {
			note := strings.TrimSpace(nextElem.Text())
			if note != "" {
				w.logger.Debug().Str("usageNote", note).Msg("Found usage note")
				newWord.AddUsageNote(note)
			}
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

				// Collect examples for this definition
				examples := []string{}
				li.ForEach("span.example", func(_ int, example *colly.HTMLElement) {
					exampleText := strings.TrimSpace(example.Text)
					if exampleText != "" {
						// Clean up the example text
						exampleText = strings.ReplaceAll(exampleText, "« ", "")
						exampleText = strings.ReplaceAll(exampleText, " »", "")
						w.logger.Debug().Str("example", exampleText).Msg("Found example")
						examples = append(examples, exampleText)
					}
				})

				// Add the definition using the entity method
				newWord.Definitions = append(newWord.Definitions, wordDomain.Definition{
					Text:     definitionText,
					WordType: "noun",
					Examples: examples,
				})
				w.logger.Debug().Int("index", len(newWord.Definitions)-1).Str("definition", definitionText).Msg("Found definition")

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
					newWord.AddSynonym(synonym)
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
					newWord.AddAntonym(antonym)
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
						newWord.Definitions = append(newWord.Definitions, wordDomain.Definition{
							Text:     definitionText,
							WordType: "",
							Examples: []string{},
						})
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
							newWord.Definitions = append(newWord.Definitions, wordDomain.Definition{
								Text:     fullText,
								WordType: "",
								Examples: []string{},
							})
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
		return nil, fmt.Errorf("no word data found: %w", wordDomain.ErrWordNotFound)
	}

	// Remove duplicate examples, synonyms, and antonyms
	newWord.Examples = removeDuplicates(newWord.Examples)
	newWord.Synonyms = removeDuplicates(newWord.Synonyms)
	newWord.Antonyms = removeDuplicates(newWord.Antonyms)

	// We can't use removeDuplicates for Definitions as it's a struct slice
	// Instead, we'll manually remove duplicates if needed

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
