package word

import (
	"time"
)

// Form represents a different form of a word with its attributes
type Form struct {
	Text       string            `json:"text"`
	Attributes map[string]string `json:"attributes,omitempty"` // e.g. {"number": "plural", "tense": "past"}
	IsLemma    bool              `json:"is_lemma,omitempty"`   // Whether this form is the lemma/base form
}

// Definition represents a single definition with its type and examples
type Definition struct {
	Text             string            `json:"text"`
	WordType         string            `json:"word_type,omitempty"` // noun, verb, adjective, etc.
	Examples         []string          `json:"examples,omitempty"`
	Register         string            `json:"register,omitempty"` // formal, informal, slang, etc.
	Gender           string            `json:"gender,omitempty"`
	Prononciation    string            `json:"prononciation,omitempty"`
	LangageSpecifics map[string]string `json:"language_specifics,omitempty"`
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
	WordType      string            `json:"word_type,omitempty"` // Primary word type if multiple exist
	Gender        string            `json:"gender,omitempty"`    // Primary gender if multiple exist
	Forms         []Form            `json:"forms,omitempty"`
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
		Forms:         []Form{},
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

// AddWordForm adds a new form of the word and updates search terms
func (w *Word) AddWordForm(text string, attributes map[string]string, isLemma bool) {
	// Add the form to WordForms
	w.Forms = append(w.Forms, Form{
		Text:       text,
		Attributes: attributes,
		IsLemma:    isLemma,
	})

	// Add to search terms if not already present
	found := false
	for _, term := range w.SearchTerms {
		if term == text {
			found = true
			break
		}
	}
	if !found {
		w.SearchTerms = append(w.SearchTerms, text)
	}

	// Update lemma if this form is marked as lemma
	if isLemma {
		w.Lemma = text
	}

	// Update the UpdatedAt timestamp
	w.UpdatedAt = time.Now()
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
