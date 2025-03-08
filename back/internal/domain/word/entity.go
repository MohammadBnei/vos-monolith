package word

import (
	"time"
)

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
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
