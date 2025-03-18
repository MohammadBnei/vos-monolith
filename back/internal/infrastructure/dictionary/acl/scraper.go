package acl

import "context"

// WiktionaryScraper defines the interface for fetching raw data from Wiktionary
type WiktionaryScraper interface {
	// FetchWordData retrieves raw word data from Wiktionary
	FetchWordData(ctx context.Context, text, language string) (*WiktionaryResponse, error)
	
	// FetchRelatedWordsData retrieves raw related words data
	FetchRelatedWordsData(ctx context.Context, word, language string) (*WiktionaryRelatedResponse, error)
	
	// FetchSuggestionsData retrieves raw suggestions data
	FetchSuggestionsData(ctx context.Context, prefix, language string) ([]string, error)
}
