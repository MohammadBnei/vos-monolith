package word

import (
	"context"
)

// DictionaryAPI defines the interface for external dictionary APIs
type DictionaryAPI interface {
	// FetchWord retrieves word information from an external API
	FetchWord(ctx context.Context, text, language string) (*Word, error)

	// FetchRelatedWords retrieves words related to the given word
	FetchRelatedWords(ctx context.Context, word *Word) (*RelatedWords, error)

	// FetchSuggestions retrieves suggestions for a given prefix and language
	FetchSuggestions(ctx context.Context, prefix, language string) ([]string, error)
}
