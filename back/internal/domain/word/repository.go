package word

import (
	"context"
)

// Repository defines the interface for word data access
type Repository interface {
	// FindByID retrieves a word by its ID
	FindByID(ctx context.Context, id string) (*Word, error)
	
	// FindByText retrieves a word by its text and language
	FindByText(ctx context.Context, text, language string) (*Word, error)

	// FindByAnyForm retrieves a word by any of its forms (using search terms)
	FindByAnyForm(ctx context.Context, text, language string) (*Word, error)

	// Save stores a word in the repository
	Save(ctx context.Context, word *Word) error

	// List retrieves words with optional filtering
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*Word, error)

	// FindByPrefix retrieves words by prefix and language
	FindByPrefix(ctx context.Context, prefix, language string, limit int) ([]*Word, error)
	
	// FindSuggestions retrieves word suggestions based on a prefix
	FindSuggestions(ctx context.Context, prefix, language string, limit int) ([]string, error)
}
