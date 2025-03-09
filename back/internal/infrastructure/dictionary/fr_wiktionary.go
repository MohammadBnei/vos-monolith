package dictionary

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
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

	// Create a single collector for the entire operation
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

	// Initialize page structure
	pageStructure := &PageStructure{
		SectionIDs:       make(map[string]string),
		HasFrenchSection: false,
		WordTypeSections: make(map[string]string),
		OtherSections:    make(map[string]string),
	}

	// Track seen examples to avoid duplicates
	seenExamples := make(map[string]bool)

	// First callback: Extract page structure from TOC
	c.OnHTML("#mw-panel-toc-list", func(e *colly.HTMLElement) {
		w.logger.Debug().Msg("Found table of contents")

		// Find the French section in the TOC
		e.ForEach("li", func(_ int, li *colly.HTMLElement) {
			sectionID := li.Attr("id")
			if sectionID == "toc-mw-content-text" {
				return
			}

			sectionTitle := strings.ReplaceAll(li.ChildAttr(".vector-toc-link", "href"), "#", "")

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
						subSectionTitle := strings.ReplaceAll(subLi.ChildAttr(".vector-toc-link", "href"), "#", "")

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
							}

							w.logger.Debug().Str("title", subSectionTitle).Str("id", actualSubID).Msg("Found subsection")

							// Parse deeper levels (definitions, synonyms, etc.)
							subLi.ForEach("ul li", func(_ int, subSubLi *colly.HTMLElement) {
								subSubSectionID := subSubLi.Attr("id")
								subSubSectionTitle := strings.ReplaceAll(subSubLi.ChildAttr(".vector-toc-link", "href"), "#", "")

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

	// Fallback for page structure if no TOC is found
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

	// Extract word forms
	c.OnHTML("table.flextable-fr-mfsp", func(e *colly.HTMLElement) {
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
		// Only set if we haven't already found it in a section
		if newWord.Pronunciation == nil || newWord.Pronunciation["ipa"] == "" {
			pronunciation := strings.TrimSpace(e.Text)
			if strings.HasPrefix(pronunciation, "\\") && strings.HasSuffix(pronunciation, "\\") {
				w.logger.Debug().Str("pronunciation", pronunciation).Msg("Found pronunciation")
				if newWord.Pronunciation == nil {
					newWord.Pronunciation = make(map[string]string)
				}
				newWord.Pronunciation["ipa"] = pronunciation
			}
		}
	})

	c.OnScraped(func(r *colly.Response) {
		// Extract etymology if section exists
		if etymologyID, ok := pageStructure.OtherSections["etymology"]; ok {
			// Use jQuery-like selector to find the etymology section
			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(r.Body))
			if err != nil {
				w.logger.Error().Err(err).Msg("Failed to parse HTML for etymology")
				return
			}

			selector := fmt.Sprintf("#%s", etymologyID)
			etymologyHeading := doc.Find(selector)
			if etymologyHeading.Length() > 0 {
				// Get the content after the etymology heading
				etymologyText := ""
				nextElem := etymologyHeading.Parent().Next()
				if nextElem.Is("dl") {
					etymologyText = strings.TrimSpace(nextElem.Text())
				} else if nextElem.Is("p") {
					etymologyText = strings.TrimSpace(nextElem.Text())
				}

				if etymologyText != "" {
					// Remove "Étymologie manquante ou incomplète" and anything after it
					if idx := strings.Index(etymologyText, "Étymologie manquante ou incomplète"); idx >= 0 {
						// If there's content before the message, keep only that part
						if idx > 0 {
							etymologyText = strings.TrimSpace(etymologyText[:idx])
						} else {
							// If the message is at the beginning, there's no useful etymology
							etymologyText = ""
						}
					}

					// Only set etymology if there's still content after filtering
					if etymologyText != "" {
						w.logger.Debug().Str("etymology", etymologyText).Msg("Found etymology")
						newWord.Etymology = cleanEtymology(etymologyText)
					} else {
						w.logger.Debug().Msg("Etymology was only the 'missing' message, ignoring")
					}
				}
			}
		}

		// Extract definitions from word type sections
		for sectionTitle, sectionID := range pageStructure.WordTypeSections {
			wordType := w.determineWordType(sectionTitle)
			foundDefinition := wordDomain.NewDefinition()
			foundDefinition.WordType = wordType

			// Use jQuery-like selector to find the definition section
			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(r.Body))
			if err != nil {
				w.logger.Error().Err(err).Msg("Failed to parse HTML for definitions")
				continue
			}

			selector := fmt.Sprintf("#%s", sectionID)
			sectionHeading := doc.Find(selector)
			if sectionHeading.Length() > 0 {
				w.logger.Debug().Str("selector", selector).Str("wordType", wordType).Msg("Found word type section")

				// Look for the ordered list that contains definitions
				var definitionList *goquery.Selection
				nextElem := sectionHeading.Parent().Next()

				// Try to find the definition list (could be directly after the heading or after a paragraph)
				for i := 0; i < 10 && nextElem.Length() > 0; i++ {
					switch {
					case nextElem.Is("p"):
						nextElem.Children().Each(func(_ int, child *goquery.Selection) {
							if titleAttr, _ := child.Attr("title"); child.Is("a") && strings.Contains(titleAttr, "Prononciation") {
								foundDefinition.Prononciation = child.Text()
							}
							if child.Is("span.ligne-de-forme") {
								foundDefinition.Gender = child.Text()
							}
						})
					case nextElem.Is("table"):
						nextElem.Find("tr").Each(func(i int, trSelection *goquery.Selection) {
							if i > 0 {
								trSelection.Find("th").Each(func(j int, tdSelection *goquery.Selection) {
									// Found masculin or feminin
									switch tdSelection.Text() {
									case "Féminin":
										foundDefinition.LangageSpecifics["feminin"] = tdSelection.Next().Text()
									case "Masculin":
										foundDefinition.LangageSpecifics["masculin"] = tdSelection.Next().Text()
									}
								})
							}
						})
					case nextElem.Is("ol"):
						definitionList = nextElem
					}
					nextElem = nextElem.Next()
				}

				if definitionList != nil {
					// Process each list item as a definition
					definitionList.ChildrenFiltered("li").Each(func(i int, liSelection *goquery.Selection) {
						// Get the main definition text
						definitionText := strings.TrimSpace(liSelection.Contents().Not("ul").Text())

						// If empty, try getting the full text and cleaning it
						if definitionText == "" {
							fullText := strings.TrimSpace(liSelection.Text())

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

						// Collect examples for this definition
						examples := []string{}
						liSelection.Find("span.example").Each(func(_ int, exampleSpan *goquery.Selection) {
							exampleText := strings.TrimSpace(exampleSpan.Text())
							if exampleText != "" {
								// Clean up the example text
								exampleText = strings.ReplaceAll(exampleText, "« ", "")
								exampleText = strings.ReplaceAll(exampleText, " »", "")

								// Skip examples with "Exemple d'utilisation manquant"
								if strings.Contains(exampleText, "utilisation manquant. (Ajouter)") {
									w.logger.Debug().Str("example", exampleText).Msg("Skipping 'missing example' message")
									return
								}

								// Only add if we haven't seen this example before
								if !seenExamples[exampleText] {
									w.logger.Debug().Str("example", exampleText).Msg("Found example")
									examples = append(examples, exampleText)
									seenExamples[exampleText] = true

									// Also add to the word's general examples
									newWord.Examples = append(newWord.Examples, exampleText)
								}
							}
						})

						foundDefinition.Text = definitionText
						foundDefinition.Examples = examples

						// Add the definition using the entity method
						newWord.Definitions = append(newWord.Definitions, foundDefinition)
						w.logger.Debug().Int("index", len(newWord.Definitions)-1).Str("definition", definitionText).Msg("Found definition")
					})
				}
			}
		}

		// Extract synonyms if section exists
		if synonymsID, ok := pageStructure.OtherSections["synonyms"]; ok {
			// Use jQuery-like selector to find the synonyms section
			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(r.Body))
			if err != nil {
				w.logger.Error().Err(err).Msg("Failed to parse HTML for synonyms")
				return
			}

			selector := fmt.Sprintf("#%s", synonymsID)
			synonymsHeading := doc.Find(selector)
			if synonymsHeading.Length() > 0 {
				w.logger.Debug().Str("selector", selector).Msg("Found synonyms section")

				// Look for the unordered list that contains synonyms
				var synonymList *goquery.Selection

				// Look for the next ul
				synonymList = synonymsHeading.Parent().Next()
				if !synonymList.Is("ul") {
					// Try the next element
					synonymList = synonymList.Next()
					if !synonymList.Is("ul") {
						return
					}
				}

				// Use goquery to iterate through list items
				synonymList.Find("li").Each(func(_ int, liSelection *goquery.Selection) {
					synonym := strings.TrimSpace(liSelection.Text())
					if synonym != "" {
						w.logger.Debug().Str("synonym", synonym).Msg("Found synonym")
						newWord.AddSynonym(synonym)
					}
				})
			}
		}

		// Extract antonyms if section exists
		if antonymsID, ok := pageStructure.OtherSections["antonyms"]; ok {
			// Use jQuery-like selector to find the antonyms section
			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(r.Body))
			if err != nil {
				w.logger.Error().Err(err).Msg("Failed to parse HTML for antonyms")
				return
			}

			selector := fmt.Sprintf("#%s", antonymsID)
			antonymsHeading := doc.Find(selector)
			if antonymsHeading.Length() > 0 {
				w.logger.Debug().Str("selector", selector).Msg("Found antonyms section")

				// Look for the unordered list that contains antonyms
				var antonymList *goquery.Selection

				// Look for the next ul
				antonymList = antonymsHeading.Parent().Next()
				if !antonymList.Is("ul") {
					// Try the next element
					antonymList = antonymList.Next()
					if !antonymList.Is("ul") {
						return
					}
				}

				// Use goquery to iterate through list items
				antonymList.Find("li").Each(func(_ int, liSelection *goquery.Selection) {
					antonym := strings.TrimSpace(liSelection.Text())
					if antonym != "" {
						w.logger.Debug().Str("antonym", antonym).Msg("Found antonym")
						newWord.AddAntonym(antonym)
					}
				})
			}
		}

		// Extract translations if section exists
		if translationsID, ok := pageStructure.OtherSections["translations"]; ok {
			// Use jQuery-like selector to find the translations section
			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(r.Body))
			if err != nil {
				w.logger.Error().Err(err).Msg("Failed to parse HTML for translations")
				return
			}

			selector := fmt.Sprintf("#%s", translationsID)
			translationsHeading := doc.Find(selector)
			if translationsHeading.Length() > 0 {
				w.logger.Debug().Str("selector", selector).Msg("Found translations section")

				// Translations are in a complex structure, navigate to find them
				var translationsDiv *goquery.Selection

				// Look for the next div with class boite
				translationsDiv = translationsHeading.Parent().Next()
				for i := 0; i < 3 && translationsDiv.Length() > 0; i++ {
					if translationsDiv.HasClass("boite") {
						break
					}
					translationsDiv = translationsDiv.Next()
				}

				if translationsDiv.HasClass("boite") {
					// Track which languages we've already processed
					processedLangs := make(map[string]bool)

					// Use goquery to iterate through list items
					translationsDiv.Find("li").Each(func(_ int, liSelection *goquery.Selection) {
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
						fullText := strings.TrimSpace(liSelection.Text())

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
							newWord.Translations[langCode] = translationText
						}
					})
				}
			}
		}

		// Fallback for definitions if none were found
		if len(newWord.Definitions) == 0 {
			w.logger.Warn().Msg("No definitions found with primary selectors, trying fallback")

			// Try to find any ordered list in the French section
			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(r.Body))
			if err != nil {
				w.logger.Error().Err(err).Msg("Failed to parse HTML for fallback definitions")
				return
			}

			doc.Find("ol li").Each(func(_ int, li *goquery.Selection) {
				if len(newWord.Definitions) == 0 {
					definitionText := strings.TrimSpace(li.Text())
					if definitionText != "" && len(definitionText) > 10 {
						w.logger.Debug().Str("definition", definitionText).Msg("Found definition with fallback method")
						newWord.Definitions = append(newWord.Definitions, wordDomain.Definition{
							Text:     definitionText,
							WordType: "",
							Examples: []string{},
						})
					}
				}
			})

			// If still no definitions, try paragraphs
			if len(newWord.Definitions) == 0 {
				doc.Find("p").Each(func(_ int, p *goquery.Selection) {
					if len(newWord.Definitions) == 0 {
						fullText := strings.TrimSpace(p.Text())
						if fullText != "" && len(fullText) > 10 && !strings.HasPrefix(fullText, "From") {
							w.logger.Debug().Str("definition", fullText).Msg("Found definition in paragraph with fallback method")
							newWord.Definitions = append(newWord.Definitions, wordDomain.Definition{
								Text:     fullText,
								WordType: "",
								Examples: []string{},
							})
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
		// Visit the page (single HTTP request)
		err := c.Visit(url)
		if err != nil {
			w.logger.Error().Err(err).Str("url", url).Msg("Failed to visit page")
			return nil, fmt.Errorf("failed to visit page: %w", err)
		}
	}

	// Wait until scraping is finished
	c.Wait()

	// If no French section was found, return an error
	if !pageStructure.HasFrenchSection {
		w.logger.Warn().Str("text", text).Str("language", language).Msg("No French section found")
		return nil, fmt.Errorf("no French section found: %w", wordDomain.ErrWordNotFound)
	}

	// If still no definitions, return error
	if len(newWord.Definitions) == 0 {
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

func cleanEtymology(etymologyText string) string {
	etymologyText = strings.TrimSpace(etymologyText)
	regex := regexp.MustCompile(`\[\d+\] `)
	etymologyText = regex.ReplaceAllString(etymologyText, "")

	return etymologyText
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
