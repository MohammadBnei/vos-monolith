package word

import (
	"context"
)

// Repository defines the interface for word data access
type Repository interface {
	// FindByText retrieves a word by its text and language
	FindByText(ctx context.Context, text, language string) (*Word, error)

	// FindByAnyForm retrieves a word by any of its forms (using search terms)
	FindByAnyForm(ctx context.Context, text, language string) (*Word, error)

	// Save stores a word in the repository
	Save(ctx context.Context, word *Word) error

	// List retrieves words with optional filtering
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*Word, error)
}
