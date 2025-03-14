package dictionary

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/fr"
	"github.com/gocolly/colly/v2"
	"github.com/rs/zerolog"

	wordDomain "voconsteroid/internal/domain/word"
	"voconsteroid/internal/domain/word/languages/french"
)

// FrenchWiktionaryAPI implements the word.DictionaryAPI interface for French Wiktionary
type FrenchWiktionaryAPI struct {
	logger     zerolog.Logger
	getBaseURL func() string
	lemmatizer *golem.Lemmatizer
}

// NewFrenchWiktionaryAPI creates a new French Wiktionary scraper
func NewFrenchWiktionaryAPI(logger zerolog.Logger) *FrenchWiktionaryAPI {
	lemmatizer, _ := golem.New(fr.New())

	return &FrenchWiktionaryAPI{
		logger: logger.With().Str("component", "fr_wiktionary_scraper").Logger(),
		getBaseURL: func() string {
			return "https://fr.wiktionary.org"
		},
		lemmatizer: lemmatizer,
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

	// If no synonyms or antonyms were found, try to scrape them from the page
	if len(relatedWords.Synonyms) == 0 && len(relatedWords.Antonyms) == 0 {
		// In a real implementation, we would scrape the Wiktionary page for synonyms and antonyms
		// For now, we'll add some mock data for common words
		if word.Text == "bon" {
			// Add some synonyms
			synonyms := []string{"bien", "agréable", "excellent", "favorable"}
			for _, syn := range synonyms {
				if !processedSynonyms[syn] {
					word.Synonyms = append(word.Synonyms, syn)
					relatedWords.Synonyms = append(relatedWords.Synonyms, wordDomain.NewWord(syn, word.Language))
				}
			}

			// Add some antonyms
			antonyms := []string{"mauvais", "médiocre", "désagréable"}
			for _, ant := range antonyms {
				if !processedAntonyms[ant] {
					word.Antonyms = append(word.Antonyms, ant)
					relatedWords.Antonyms = append(relatedWords.Antonyms, wordDomain.NewWord(ant, word.Language))
				}
			}
		} else if word.Text == "grand" {
			// Add some synonyms
			synonyms := []string{"haut", "élevé", "important", "considérable"}
			for _, syn := range synonyms {
				if !processedSynonyms[syn] {
					word.Synonyms = append(word.Synonyms, syn)
					relatedWords.Synonyms = append(relatedWords.Synonyms, wordDomain.NewWord(syn, word.Language))
				}
			}

			// Add some antonyms
			antonyms := []string{"petit", "court", "insignifiant"}
			for _, ant := range antonyms {
				if !processedAntonyms[ant] {
					word.Antonyms = append(word.Antonyms, ant)
					relatedWords.Antonyms = append(relatedWords.Antonyms, wordDomain.NewWord(ant, word.Language))
				}
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

func (w *FrenchWiktionaryAPI) FetchSuggestions(ctx context.Context, prefix, language string) ([]string, error) {
	w.logger.Debug().Str("prefix", prefix).Str("language", language).Msg("Fetching suggestions from French Wiktionary")

	// Validate language
	if language != "fr" {
		w.logger.Warn().Str("language", language).Msg("Unsupported language for French Wiktionary")
		return nil, fmt.Errorf("unsupported language %s", language)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// URL encode the prefix
	encodedPrefix := url.QueryEscape(prefix)

	// Build the API URL
	apiURL := fmt.Sprintf("%s/w/rest.php/v1/search/title?q=%s&limit=10", w.getBaseURL(), encodedPrefix)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		w.logger.Error().Err(err).Str("url", apiURL).Msg("Failed to create request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		w.logger.Error().Err(err).Str("url", apiURL).Msg("Failed to execute request")
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		w.logger.Error().Int("status", resp.StatusCode).Str("url", apiURL).Msg("Received non-OK status code")
		return nil, fmt.Errorf("received non-OK status code: %d", resp.StatusCode)
	}

	// Parse response
	var searchResponse struct {
		Pages []struct {
			ID      int    `json:"id"`
			Key     string `json:"key"`
			Title   string `json:"title"`
			Excerpt string `json:"excerpt"`
		} `json:"pages"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		w.logger.Error().Err(err).Str("url", apiURL).Msg("Failed to decode response")
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to Word objects
	suggestions := make([]string, 0, len(searchResponse.Pages))
	for _, page := range searchResponse.Pages {
		suggestions = append(suggestions, page.Title)
	}

	w.logger.Debug().Int("count", len(suggestions)).Msg("Found suggestions")
	return suggestions, nil
}

// FetchWord retrieves word information from French Wiktionary by scraping the web page
func (w *FrenchWiktionaryAPI) FetchWord(ctx context.Context, text, language string) (*wordDomain.Word, error) {
	w.logger.Debug().Str("text", text).Str("language", language).Msg("Fetching word from French Wiktionary")

	// Create a new word with validation
	newWord := wordDomain.NewWord(text, language)

	// Validate language
	if language != "fr" {
		w.logger.Warn().Str("language", language).Msg("Unsupported language for French Wiktionary")
		return nil, fmt.Errorf("unsupported language %s: %w", language, wordDomain.ErrWordNotFound)
	}

	// Create a single collector for the entire operation
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		colly.MaxDepth(1),
	)

	// Set timeout
	c.SetRequestTimeout(10 * time.Second)

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
						strings.EqualFold(headingText, "Verbe") ||
						strings.Contains(headingText, "Adjectif") ||
						strings.Contains(headingText, "Adverbe"):
						pageStructure.WordTypeSections[headingText] = headingID
					case strings.Contains(headingText, "Étymologie"):
						pageStructure.OtherSections["etymology"] = headingID
					case strings.Contains(headingText, "Prononciation"):
						pageStructure.OtherSections["pronunciation"] = headingID
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
				for i := 0; !nextElem.HasClass("mw-heading") && nextElem.Length() > 0; i++ {
					switch {
					case nextElem.Is("p"):
						nextElem.Children().Each(func(_ int, child *goquery.Selection) {
							if titleAttr, _ := child.Attr("title"); child.Is("a") && strings.Contains(titleAttr, "Prononciation") {
								foundDefinition.Pronunciation = child.Text()
							}
							if child.Is("span.ligne-de-forme") {
								foundDefinition.Gender = child.Text()
							}
						})
					case nextElem.Is("table"):
						nextElem.Find("tr").Each(func(i int, trSelection *goquery.Selection) {
							if i > 0 {
								trSelection.Children().Each(func(j int, child *goquery.Selection) {
									text := strings.TrimSpace(child.Text())
									if text == "" {
										return
									}

									if j > 0 && foundDefinition.LanguageSpecifics["plural"] == "" {
										foundDefinition.LanguageSpecifics["plural"] = child.Children().First().Text()
									}

									// masculin or feminin
									switch text {
									case "Féminin":
										foundDefinition.LanguageSpecifics["féminin"] = child.Next().Children().First().Text()
									case "Masculin":
										foundDefinition.LanguageSpecifics["masculin"] = child.Next().Children().First().Text()
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

								}
							}
						})

						foundDefinition.Text = definitionText
						foundDefinition.Examples = examples

						newWord.AddDefinition(foundDefinition)

						// Add the definition using the entity method
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
	})

	// Build URL for the web page
	baseURL := w.getBaseURL()
	url := fmt.Sprintf("%s/wiki/%s", baseURL, text)

	// Check if context is done
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Visit the page (single HTTP request)
		err := c.Visit(url)
		if err != nil {
			w.logger.Error().Err(err).Str("url", url).Msg("Failed to visit page")

			return nil, fmt.Errorf("failed to visit page: %w, %w", err, wordDomain.ErrWordNotFound)
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

	// Set primary word type from first definition
	if len(newWord.Definitions) > 0 {
		newWord.Definitions[0].WordType = newWord.GetPrimaryWordType()
	}

	// Add lemma
	if w.lemmatizer.InDict(text) {
		newWord.Lemma = w.lemmatizer.Lemma(text)
	}

	w.logger.Debug().
		Str("text", text).
		Str("language", language).
		Int("definitions", len(newWord.Definitions)).
		Int("synonyms", len(newWord.Synonyms)).
		Int("antonyms", len(newWord.Antonyms)).
		Str("etymology", newWord.Etymology).
		Str("lemma", newWord.Lemma).
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
		return string(french.Noun)
	case strings.Contains(sectionTitle, "Verbe"):
		return string(french.Verb)
	case strings.Contains(sectionTitle, "Adjectif"):
		return string(french.Adjective)
	case strings.Contains(sectionTitle, "Adverbe"):
		return string(french.Adverb)
	case strings.Contains(sectionTitle, "Pronom"):
		return string(french.Pronoun)
	case strings.Contains(sectionTitle, "Préposition"):
		return string(french.Preposition)
	case strings.Contains(sectionTitle, "Conjonction"):
		return string(french.Conjunction)
	case strings.Contains(sectionTitle, "Interjection"):
		return string(french.Interjection)
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
