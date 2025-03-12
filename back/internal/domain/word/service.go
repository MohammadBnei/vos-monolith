package word

import (
	"context"
)

// Service defines the interface for word business logic
type Service interface {
	// Search finds a word by text and language, fetching from external API if needed
	Search(ctx context.Context, text, language string) (*Word, error)

	// GetRecentWords retrieves recently searched words
	GetRecentWords(ctx context.Context, language string, limit int) ([]*Word, error)

	// GetRelatedWords finds words related to the given word (synonyms, antonyms)
	GetRelatedWords(ctx context.Context, wordID string) (*RelatedWords, error)

	// AutoComplete provides autocomplete suggestions based on prefix and language
	AutoComplete(ctx context.Context, prefix, language string) ([]string, error)

	// GetSuggestions retrieves word suggestions based on a prefix
	GetSuggestions(ctx context.Context, prefix, language string) ([]string, error)
}

// RelatedWords groups words related to a specific word
type RelatedWords struct {
	SourceWord *Word   `json:"source_word"`
	Synonyms   []*Word `json:"synonyms,omitempty"`
	Antonyms   []*Word `json:"antonyms,omitempty"`
}
