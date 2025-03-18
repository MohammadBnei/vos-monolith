package word

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"voconsteroid/internal/domain/word/languages/english"
	"voconsteroid/internal/domain/word/languages/french"
)

// Translation represents a value object for storing translations in a target language
type Translation struct {
	Language string   `json:"language"` // ISO 639-1 code
	Terms    []string `json:"terms"`    // List of translated terms
}

// Definition represents a single definition with its type and examples
type Definition struct {
	ID                string            `json:"id"`
	Text              string            `json:"text"`
	WordType          string            `json:"word_type,omitempty"` // noun, verb, adjective, etc.
	Examples          []string          `json:"examples,omitempty"`
	Gender            string            `json:"gender,omitempty"`
	Pronunciation     string            `json:"pronunciation,omitempty"`
	LanguageSpecifics map[string]string `json:"language_specifics,omitempty"`
	Notes             []string          `json:"notes,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

// Word represents a vocabulary word with its definitions and metadata
type Word struct {
	ID           string              `json:"id"`
	Text         string              `json:"text"` // The canonical form
	Language     string              `json:"language"`
	Definitions  []Definition        `json:"definitions,omitempty"`
	Pronunciation string             `json:"pronunciation,omitempty"` // Word-level pronunciation
	Etymology    string              `json:"etymology,omitempty"`
	Translations map[string][]string `json:"translations,omitempty"` // Map of language codes to lists of translated terms
	Synonyms     []string            `json:"synonyms,omitempty"`
	Antonyms     []string            `json:"antonyms,omitempty"`
	SearchTerms  []string            `json:"search_terms,omitempty"` // All searchable forms of the word
	Lemma        string              `json:"lemma,omitempty"`        // Base form of the word
	UsageNotes   []string            `json:"usage_notes,omitempty"`  // General usage information
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

// NewWord creates a new Word entity
func NewWord(text, language string) *Word {
	now := time.Now()
	id := uuid.New().String()
	return &Word{
		ID:           id,
		Text:         text,
		Language:     language,
		Definitions:  []Definition{},
		Synonyms:     []string{},
		Antonyms:     []string{},
		Translations: make(map[string][]string),
		SearchTerms:  []string{text}, // Initialize with the main text as a search term
		UsageNotes:   []string{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// NewDefinition returns a new, empty Definition structure.
func NewDefinition() Definition {
	now := time.Now()
	return Definition{
		ID:                uuid.New().String(),
		Examples:          []string{},
		LanguageSpecifics: make(map[string]string),
		Notes:             []string{},
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// AddExample adds an example to the definition
func (d *Definition) AddExample(example string) {
	// Check if example already exists
	for _, ex := range d.Examples {
		if ex == example {
			return
		}
	}
	
	d.Examples = append(d.Examples, example)
	d.UpdatedAt = time.Now()
}

// RemoveExample removes an example by index
func (d *Definition) RemoveExample(index int) error {
	if index < 0 || index >= len(d.Examples) {
		return fmt.Errorf("index out of range")
	}
	
	d.Examples = append(d.Examples[:index], d.Examples[index+1:]...)
	d.UpdatedAt = time.Now()
	return nil
}

// SetWordType sets and validates the word type
func (d *Definition) SetWordType(wordType string) {
	d.WordType = wordType
	d.UpdatedAt = time.Now()
}

// SetGender sets the gender
func (d *Definition) SetGender(gender string) {
	d.Gender = gender
	d.UpdatedAt = time.Now()
}

// SetPronunciation sets the pronunciation
func (d *Definition) SetPronunciation(pronunciation string) {
	d.Pronunciation = pronunciation
	d.UpdatedAt = time.Now()
}

// AddLanguageSpecific adds a language-specific detail
func (d *Definition) AddLanguageSpecific(key, value string) {
	d.LanguageSpecifics[key] = value
	d.UpdatedAt = time.Now()
}

// RemoveLanguageSpecific removes a language-specific detail
// Returns true if the key was found and removed
func (d *Definition) RemoveLanguageSpecific(key string) bool {
	_, exists := d.LanguageSpecifics[key]
	if exists {
		delete(d.LanguageSpecifics, key)
		d.UpdatedAt = time.Now()
		return true
	}
	return false
}

// AddNote adds a note to the definition
func (d *Definition) AddNote(note string) {
	d.Notes = append(d.Notes, note)
	d.UpdatedAt = time.Now()
}

// RemoveNote removes a note by index
func (d *Definition) RemoveNote(index int) error {
	if index < 0 || index >= len(d.Notes) {
		return fmt.Errorf("index out of range")
	}
	
	d.Notes = append(d.Notes[:index], d.Notes[index+1:]...)
	d.UpdatedAt = time.Now()
	return nil
}

// Update updates this definition with values from another
func (d *Definition) Update(updatedDef Definition) {
	// Preserve ID and creation timestamp
	id := d.ID
	createdAt := d.CreatedAt
	
	// Copy all fields from the updated definition
	*d = updatedDef
	
	// Restore preserved fields
	d.ID = id
	d.CreatedAt = createdAt
	d.UpdatedAt = time.Now()
}

// NewTranslation creates a new Translation value object
func NewTranslation(language string, terms []string) Translation {
	return Translation{
		Language: language,
		Terms:    terms,
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
	// Ensure the definition has an ID
	if definition.ID == "" {
		definition.ID = uuid.New().String()
		definition.CreatedAt = time.Now()
		definition.UpdatedAt = time.Now()
	}
	
	w.Definitions = append(w.Definitions, definition)

	// Update the search terms
	for _, term := range definition.LanguageSpecifics {
		w.AddSearchTerm(term)
	}

	// Update the UpdatedAt timestamp
	w.UpdatedAt = time.Now()
}

// UpdateDefinition updates an existing definition by ID
func (w *Word) UpdateDefinition(definitionID string, updatedDef Definition) error {
	for i, def := range w.Definitions {
		if def.ID == definitionID {
			// Preserve the original ID, CreatedAt
			updatedDef.ID = def.ID
			updatedDef.CreatedAt = def.CreatedAt
			updatedDef.UpdatedAt = time.Now()
			
			// Replace the definition
			w.Definitions[i] = updatedDef
			
			// Update search terms
			for _, term := range updatedDef.LanguageSpecifics {
				w.AddSearchTerm(term)
			}
			
			w.UpdatedAt = time.Now()
			return nil
		}
	}
	
	return fmt.Errorf("definition with ID %s not found", definitionID)
}

// RemoveDefinition removes a definition by ID
func (w *Word) RemoveDefinition(definitionID string) error {
	for i, def := range w.Definitions {
		if def.ID == definitionID {
			w.Definitions = append(w.Definitions[:i], w.Definitions[i+1:]...)
			w.UpdatedAt = time.Now()
			return nil
		}
	}
	
	return fmt.Errorf("definition with ID %s not found", definitionID)
}

// AddExample adds an example to a specific definition
func (w *Word) AddExample(definitionID string, example string) error {
	for i, def := range w.Definitions {
		if def.ID == definitionID {
			// Check if example already exists
			for _, ex := range def.Examples {
				if ex == example {
					return nil // Example already exists
				}
			}
			
			// Add the example
			w.Definitions[i].Examples = append(def.Examples, example)
			w.Definitions[i].UpdatedAt = time.Now()
			w.UpdatedAt = time.Now()
			return nil
		}
	}
	
	return fmt.Errorf("definition with ID %s not found", definitionID)
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

// RemoveSynonym removes a synonym and returns true if it was found and removed
func (w *Word) RemoveSynonym(synonym string) bool {
	for i, s := range w.Synonyms {
		if s == synonym {
			w.Synonyms = append(w.Synonyms[:i], w.Synonyms[i+1:]...)
			w.UpdatedAt = time.Now()
			return true
		}
	}
	return false
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

// RemoveAntonym removes an antonym and returns true if it was found and removed
func (w *Word) RemoveAntonym(antonym string) bool {
	for i, a := range w.Antonyms {
		if a == antonym {
			w.Antonyms = append(w.Antonyms[:i], w.Antonyms[i+1:]...)
			w.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// AddTranslation adds a translation term to the specified language
func (w *Word) AddTranslation(language string, term string) {
	// Check if we already have translations for this language
	terms, exists := w.Translations[language]
	
	if !exists {
		// Create a new entry for this language
		w.Translations[language] = []string{term}
	} else {
		// Check if the term already exists
		for _, existingTerm := range terms {
			if existingTerm == term {
				return // Term already exists, no need to add
			}
		}
		// Add the new term
		w.Translations[language] = append(terms, term)
	}
	
	w.UpdatedAt = time.Now()
}

// RemoveTranslation removes a specific translation term from a language
// Returns true if the term was found and removed
func (w *Word) RemoveTranslation(language string, term string) bool {
	terms, exists := w.Translations[language]
	if !exists {
		return false
	}
	
	for i, existingTerm := range terms {
		if existingTerm == term {
			// Remove the term
			w.Translations[language] = append(terms[:i], terms[i+1:]...)
			
			// If no terms left for this language, remove the language entry
			if len(w.Translations[language]) == 0 {
				delete(w.Translations, language)
			}
			
			w.UpdatedAt = time.Now()
			return true
		}
	}
	
	return false
}

// ClearTranslations removes all translations for a specific language
// Returns true if the language existed and translations were cleared
func (w *Word) ClearTranslations(language string) bool {
	_, exists := w.Translations[language]
	if exists {
		delete(w.Translations, language)
		w.UpdatedAt = time.Now()
		return true
	}
	return false
}

// SetPronunciation sets the word-level pronunciation
func (w *Word) SetPronunciation(pronunciation string) {
	w.Pronunciation = pronunciation
	w.UpdatedAt = time.Now()
}

// SetEtymology sets the etymology of the word
func (w *Word) SetEtymology(etymology string) {
	w.Etymology = etymology
	w.UpdatedAt = time.Now()
}

// AddUsageNote adds a new usage note
func (w *Word) AddUsageNote(note string) {
	w.UsageNotes = append(w.UsageNotes, note)
	w.UpdatedAt = time.Now()
}

// RemoveUsageNote removes a usage note by index
func (w *Word) RemoveUsageNote(index int) error {
	if index < 0 || index >= len(w.UsageNotes) {
		return fmt.Errorf("index out of range")
	}
	
	// Remove the note at the specified index
	w.UsageNotes = append(w.UsageNotes[:index], w.UsageNotes[index+1:]...)
	w.UpdatedAt = time.Now()
	return nil
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

// Validate validates the entire word entity against domain rules
func (w *Word) Validate() error {
	// Check required fields
	if w.Text == "" {
		return fmt.Errorf("word text cannot be empty")
	}
	
	if w.Language == "" {
		return fmt.Errorf("language cannot be empty")
	}
	
	// Validate all definitions
	for _, def := range w.Definitions {
		if err := w.ValidateDefinition(def); err != nil {
			return err
		}
	}
	
	// Validate translations (ISO 639-1 codes)
	for lang := range w.Translations {
		if len(lang) != 2 {
			return fmt.Errorf("invalid language code: %s, must be ISO 639-1 format", lang)
		}
	}
	
	return nil
}

// MergeWith merges data from another word (for enrichment)
func (w *Word) MergeWith(other *Word) error {
	if w.Text != other.Text || w.Language != other.Language {
		return fmt.Errorf("cannot merge words with different text or language")
	}
	
	// Merge definitions (by ID)
	for _, otherDef := range other.Definitions {
		found := false
		for i, def := range w.Definitions {
			if def.ID == otherDef.ID {
				// Update existing definition
				w.Definitions[i] = otherDef
				found = true
				break
			}
		}
		
		if !found {
			// Add new definition
			w.AddDefinition(otherDef)
		}
	}
	
	// Merge etymology if empty
	if w.Etymology == "" && other.Etymology != "" {
		w.Etymology = other.Etymology
	}
	
	// Merge pronunciation if empty
	if w.Pronunciation == "" && other.Pronunciation != "" {
		w.Pronunciation = other.Pronunciation
	}
	
	// Merge translations
	for lang, terms := range other.Translations {
		for _, term := range terms {
			w.AddTranslation(lang, term)
		}
	}
	
	// Merge synonyms
	for _, syn := range other.Synonyms {
		w.AddSynonym(syn)
	}
	
	// Merge antonyms
	for _, ant := range other.Antonyms {
		w.AddAntonym(ant)
	}
	
	// Merge search terms
	for _, term := range other.SearchTerms {
		w.AddSearchTerm(term)
	}
	
	// Merge usage notes
	for _, note := range other.UsageNotes {
		w.AddUsageNote(note)
	}
	
	w.UpdatedAt = time.Now()
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
