package word

import (
	"errors"
	"time"
	
	"voconsteroid/internal/domain/word/languages/english"
	"voconsteroid/internal/domain/word/languages/french"
)

var (
	ErrInvalidWordType = errors.New("invalid word type")
	ErrInvalidGender   = errors.New("invalid gender")
)

// Definition represents a single definition with its type and examples
type Definition struct {
	Text             string            `json:"text"`
	WordType         string            `json:"word_type,omitempty"` // noun, verb, adjective, etc.
	Examples         []string          `json:"examples,omitempty"`
	Gender           string            `json:"gender,omitempty"`
	Pronunciation    string            `json:"pronunciation,omitempty"`
	LanguageSpecifics map[string]string `json:"language_specifics,omitempty"`
	Notes            []string          `json:"notes,omitempty"`
}

// Word represents a vocabulary word with its definitions and metadata
type Word struct {
	ID            string            `json:"id"`
	Text          string            `json:"text"` // The canonical form
	Language      string            `json:"language"`
	Definitions   []Definition      `json:"definitions,omitempty"`
	Examples      []string          `json:"examples,omitempty"`      // General examples not tied to a specific definition
	Pronunciation map[string]string `json:"pronunciation,omitempty"` // Different pronunciation formats (IPA, audio URL, etc.)
	Etymology     string            `json:"etymology,omitempty"`
	Translations  map[string]string `json:"translations,omitempty"`
	Synonyms      []string          `json:"synonyms,omitempty"`
	Antonyms      []string          `json:"antonyms,omitempty"`
	SearchTerms   []string          `json:"search_terms,omitempty"` // All searchable forms of the word
	Lemma         string            `json:"lemma,omitempty"`        // Base form of the word
	UsageNotes    []string          `json:"usage_notes,omitempty"`  // General usage information
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// NewWord creates a new Word entity
func NewWord(text, language string) *Word {
	now := time.Now()
	return &Word{
		Text:          text,
		Language:      language,
		Definitions:   []Definition{},
		Examples:      []string{},
		Pronunciation: make(map[string]string),
		Synonyms:      []string{},
		Antonyms:      []string{},
		Translations:  make(map[string]string),
		SearchTerms:   []string{text}, // Initialize with the main text as a search term
		UsageNotes:    []string{},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// NewDefinition returns a new, empty Definition structure.
func NewDefinition() Definition {
	return Definition{
		Examples:         []string{},
		LangageSpecifics: make(map[string]string),
		Notes:            []string{},
	}
}

// SetLemma sets the lemma (base form) of the word
func (w *Word) SetLemma(lemma string) {
	w.Lemma = lemma

	// If the lemma is not already in search terms, add it
	found := false
	for _, term := range w.SearchTerms {
		if term == lemma {
			found = true
			break
		}
	}
	if !found {
		w.SearchTerms = append(w.SearchTerms, lemma)
	}

	// Update the UpdatedAt timestamp
	w.UpdatedAt = time.Now()
}

// AddDefinition adds a new definition to the word
func (w *Word) AddDefinition(definition Definition) {
	w.Definitions = append(w.Definitions, definition)

	// Update the primary word type if not set
	if w.WordType == "" && definition.WordType != "" {
		w.WordType = definition.WordType
	}

	// Update the search terms
	for _, term := range definition.LangageSpecifics {
		w.AddSearchTerm(term)
	}

	// Update the UpdatedAt timestamp
	w.UpdatedAt = time.Now()
}

// AddSearchTerm adds a new search term to the word's search terms.
// It is used to make the word searchable by a particular term.
func (w *Word) AddSearchTerm(term string) {
	// Check if already exists
	for _, s := range w.SearchTerms {
		if s == term {
			return
		}
	}

	w.SearchTerms = append(w.SearchTerms, term)
	w.UpdatedAt = time.Now()
}

// AddSynonym adds a new synonym if not already present
func (w *Word) AddSynonym(synonym string) {
	// Check if already exists
	for _, s := range w.Synonyms {
		if s == synonym {
			return
		}
	}

	w.Synonyms = append(w.Synonyms, synonym)
	w.UpdatedAt = time.Now()
}

// AddAntonym adds a new antonym if not already present
func (w *Word) AddAntonym(antonym string) {
	// Check if already exists
	for _, a := range w.Antonyms {
		if a == antonym {
			return
		}
	}

	w.Antonyms = append(w.Antonyms, antonym)
	w.UpdatedAt = time.Now()
}

// AddUsageNote adds a new usage note
func (w *Word) AddUsageNote(note string) {
	w.UsageNotes = append(w.UsageNotes, note)
	w.UpdatedAt = time.Now()
}

// SetPronunciation sets a pronunciation in the specified format
func (w *Word) SetPronunciation(format, value string) {
	if w.Pronunciation == nil {
		w.Pronunciation = make(map[string]string)
	}
	w.Pronunciation[format] = value
	w.UpdatedAt = time.Now()
}

// ValidateDefinition validates a definition based on language rules
func (w *Word) ValidateDefinition(def Definition) error {
	switch w.Language {
	case "fr":
		if def.WordType != "" && !french.IsValidWordType(french.WordType(def.WordType)) {
			return ErrInvalidWordType
		}
		if def.Gender != "" && !french.IsValidGender(french.Gender(def.Gender)) {
			return ErrInvalidGender
		}
	case "en":
		if def.WordType != "" && !english.IsValidWordType(english.WordType(def.WordType)) {
			return ErrInvalidWordType
		}
		// English doesn't have grammatical gender
	}
	return nil
}

// GetPrimaryWordType returns the word type of the first definition
func (w *Word) GetPrimaryWordType() string {
	if len(w.Definitions) > 0 {
		return w.Definitions[0].WordType
	}
	return ""
}

// GetAllSpecifics returns all specifics of the word from all definitions
func (w *Word) GetAllSpecifics() []string {
	specifics := make([]string, 0)
	for _, def := range w.Definitions {
		for _, val := range def.LanguageSpecifics {
			specifics = append(specifics, val)
		}
	}
	return specifics
}

// GetDefinitionsByType filters definitions by type
func (w *Word) GetDefinitionsByType(wordType string) []Definition {
	defs := make([]Definition, 0)
	for _, def := range w.Definitions {
		if def.WordType == wordType {
			defs = append(defs, def)
		}
	}
	return defs
}
