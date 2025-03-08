package word

import (
	"context"
)

// DictionaryAPI defines the interface for external dictionary APIs
type DictionaryAPI interface {
	// FetchWord retrieves word information from an external API
	FetchWord(ctx context.Context, text, language string) (*Word, error)
}
