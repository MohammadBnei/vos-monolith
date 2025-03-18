package acl

import (
	"context"
	"fmt"
	
	"github.com/rs/zerolog"
	
	wordDomain "voconsteroid/internal/domain/word"
)

// WiktionaryAdapter transforms raw Wiktionary data into domain objects
type WiktionaryAdapter struct {
	logger zerolog.Logger
}

// NewWiktionaryAdapter creates a new WiktionaryAdapter
func NewWiktionaryAdapter(logger zerolog.Logger) *WiktionaryAdapter {
	return &WiktionaryAdapter{
		logger: logger.With().Str("component", "wiktionary_adapter").Logger(),
	}
}

// TransformToWord converts a WiktionaryResponse to a domain Word
func (a *WiktionaryAdapter) TransformToWord(response *WiktionaryResponse) *wordDomain.Word {
	a.logger.Debug().Str("word", response.Word).Msg("Transforming Wiktionary response to domain Word")
	
	// Create a new Word entity
	w := wordDomain.NewWord(response.Word, response.Language)
	
	// Set etymology if available
	if response.Etymology != "" {
		w.SetEtymology(response.Etymology)
	}
	
	// Set pronunciation if available
	if response.Pronunciation != "" {
		w.SetPronunciation(response.Pronunciation)
	}
	
	// Set lemma if available
	if response.Lemma != "" {
		w.SetLemma(response.Lemma)
	}
	
	// Add definitions
	for _, def := range response.Definitions {
		domainDef := wordDomain.NewDefinition()
		domainDef.Text = def.Text
		domainDef.SetWordType(def.Type)
		
		if def.Gender != "" {
			domainDef.SetGender(def.Gender)
		}
		
		if def.Pronunciation != "" {
			domainDef.SetPronunciation(def.Pronunciation)
		}
		
		// Add examples
		for _, example := range def.Examples {
			domainDef.AddExample(example)
		}
		
		// Add language specifics
		for key, value := range def.LanguageSpecifics {
			domainDef.AddLanguageSpecific(key, value)
		}
		
		// Add notes
		for _, note := range def.Notes {
			domainDef.AddNote(note)
		}
		
		w.AddDefinition(domainDef)
	}
	
	// Add translations
	for lang, terms := range response.Translations {
		for _, term := range terms {
			w.AddTranslation(lang, term)
		}
	}
	
	// Add synonyms
	for _, syn := range response.Synonyms {
		w.AddSynonym(syn)
	}
	
	// Add antonyms
	for _, ant := range response.Antonyms {
		w.AddAntonym(ant)
	}
	
	// Add search terms
	for _, term := range response.SearchTerms {
		w.AddSearchTerm(term)
	}
	
	return w
}

// TransformToRelatedWords converts a WiktionaryRelatedResponse to a domain RelatedWords
func (a *WiktionaryAdapter) TransformToRelatedWords(response *WiktionaryRelatedResponse, sourceWord *wordDomain.Word) (*wordDomain.RelatedWords, error) {
	a.logger.Debug().Str("word", response.SourceWord).Msg("Transforming related words response")
	
	// Create a new RelatedWords entity
	relatedWords := &wordDomain.RelatedWords{
		SourceWord: sourceWord,
		Synonyms:   []*wordDomain.Word{},
		Antonyms:   []*wordDomain.Word{},
	}
	
	// Add synonyms as minimal Word objects
	for _, syn := range response.Synonyms {
		synWord := wordDomain.NewWord(syn, response.Language)
		relatedWords.Synonyms = append(relatedWords.Synonyms, synWord)
	}
	
	// Add antonyms as minimal Word objects
	for _, ant := range response.Antonyms {
		antWord := wordDomain.NewWord(ant, response.Language)
		relatedWords.Antonyms = append(relatedWords.Antonyms, antWord)
	}
	
	return relatedWords, nil
}

// FetchAndTransformWord fetches a word from Wiktionary and transforms it
func (a *WiktionaryAdapter) FetchAndTransformWord(ctx context.Context, scraper WiktionaryScraper, text, language string) (*wordDomain.Word, error) {
	a.logger.Debug().Str("word", text).Str("language", language).Msg("Fetching and transforming word")
	
	// Fetch raw data from Wiktionary
	response, err := scraper.FetchWordData(ctx, text, language)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch word data: %w", err)
	}
	
	// Transform to domain object
	return a.TransformToWord(response), nil
}

// FetchAndTransformRelatedWords fetches related words and transforms them
func (a *WiktionaryAdapter) FetchAndTransformRelatedWords(ctx context.Context, scraper WiktionaryScraper, word *wordDomain.Word) (*wordDomain.RelatedWords, error) {
	a.logger.Debug().Str("word", word.Text).Str("language", word.Language).Msg("Fetching and transforming related words")
	
	// Fetch raw data from Wiktionary
	response, err := scraper.FetchRelatedWordsData(ctx, word.Text, word.Language)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch related words data: %w", err)
	}
	
	// Transform to domain object
	return a.TransformToRelatedWords(response, word)
}

// EnrichWord enriches a word with missing fields
func (a *WiktionaryAdapter) EnrichWord(ctx context.Context, scraper WiktionaryScraper, word *wordDomain.Word, status EnrichmentStatus) (*wordDomain.Word, error) {
	a.logger.Debug().Str("word", word.Text).Str("language", word.Language).Msg("Enriching word with missing fields")
	
	// Fetch raw data from Wiktionary
	response, err := scraper.FetchWordData(ctx, word.Text, word.Language)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch word data for enrichment: %w", err)
	}
	
	// Create a new Word to hold the enriched data
	enrichedWord := wordDomain.NewWord(word.Text, word.Language)
	
	// Only copy the fields that need enrichment
	if status.NeedsDefinitions {
		for _, def := range response.Definitions {
			domainDef := wordDomain.NewDefinition()
			domainDef.Text = def.Text
			domainDef.SetWordType(def.Type)
			
			if def.Gender != "" {
				domainDef.SetGender(def.Gender)
			}
			
			if def.Pronunciation != "" {
				domainDef.SetPronunciation(def.Pronunciation)
			}
			
			// Add examples
			for _, example := range def.Examples {
				domainDef.AddExample(example)
			}
			
			// Add language specifics
			for key, value := range def.LanguageSpecifics {
				domainDef.AddLanguageSpecific(key, value)
			}
			
			// Add notes
			for _, note := range def.Notes {
				domainDef.AddNote(note)
			}
			
			enrichedWord.AddDefinition(domainDef)
		}
	}
	
	if status.NeedsEtymology && response.Etymology != "" {
		enrichedWord.SetEtymology(response.Etymology)
	}
	
	if status.NeedsPronunciation && response.Pronunciation != "" {
		enrichedWord.SetPronunciation(response.Pronunciation)
	}
	
	if status.NeedsTranslations {
		for lang, terms := range response.Translations {
			for _, term := range terms {
				enrichedWord.AddTranslation(lang, term)
			}
		}
	}
	
	if status.NeedsSynonyms {
		for _, syn := range response.Synonyms {
			enrichedWord.AddSynonym(syn)
		}
	}
	
	if status.NeedsAntonyms {
		for _, ant := range response.Antonyms {
			enrichedWord.AddAntonym(ant)
		}
	}
	
	// Merge the enriched data with the original word
	err = word.MergeWith(enrichedWord)
	if err != nil {
		return nil, fmt.Errorf("failed to merge enriched data: %w", err)
	}
	
	return word, nil
}
