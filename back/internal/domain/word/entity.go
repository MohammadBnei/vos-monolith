package word

import (
	"time"
)

// WordForm represents a different form of a word with its attributes
type WordForm struct {
	Text       string            `json:"text"`
	Attributes map[string]string `json:"attributes,omitempty"` // e.g. {"number": "plural", "tense": "past"}
	IsLemma    bool              `json:"is_lemma,omitempty"`   // Whether this form is the lemma/base form
}

// Word represents a vocabulary word with its definitions and metadata
type Word struct {
	ID            string            `json:"id"`
	Text          string            `json:"text"`
	Language      string            `json:"language"`
	Definitions   []string          `json:"definitions,omitempty"`
	Examples      []string          `json:"examples,omitempty"`
	Pronunciation string            `json:"pronunciation,omitempty"`
	Etymology     string            `json:"etymology,omitempty"`
	Translations  map[string]string `json:"translations,omitempty"`
	Synonyms      []string          `json:"synonyms,omitempty"`
	WordType      string            `json:"word_type,omitempty"`
	Gender        string            `json:"gender,omitempty"`
	WordForms     []WordForm        `json:"word_forms,omitempty"`
	SearchTerms   []string          `json:"search_terms,omitempty"` // All searchable forms of the word
	Lemma         string            `json:"lemma,omitempty"`        // Base form of the word
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// NewWord creates a new Word entity
func NewWord(text, language string) *Word {
	now := time.Now()
	return &Word{
		Text:         text,
		Language:     language,
		Definitions:  []string{},
		Examples:     []string{},
		Synonyms:     []string{},
		Translations: make(map[string]string),
		WordForms:    []WordForm{},
		SearchTerms:  []string{text}, // Initialize with the main text as a search term
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// AddWordForm adds a new form of the word and updates search terms
func (w *Word) AddWordForm(text string, attributes map[string]string, isLemma bool) {
	// Add the form to WordForms
	w.WordForms = append(w.WordForms, WordForm{
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
