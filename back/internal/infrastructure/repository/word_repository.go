package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"voconsteroid/internal/domain/word"
)

// WordRepository implements the word.Repository interface using PostgreSQL
type WordRepository struct {
	db     *pgxpool.Pool
	logger zerolog.Logger
}

// NewWordRepository creates a new word repository
func NewWordRepository(db *pgxpool.Pool, logger zerolog.Logger) *WordRepository {
	return &WordRepository{
		db:     db,
		logger: logger.With().Str("component", "word_repository").Logger(),
	}
}

// FindByText retrieves a word by its text and language
func (r *WordRepository) FindByText(ctx context.Context, text, language string) (*word.Word, error) {
	query := `
		SELECT id, text, language, definitions, examples, pronunciation, etymology, translations, created_at, updated_at
		FROM words
		WHERE text = $1 AND language = $2
	`

	var w word.Word
	var definitions, examples []string
	var translations map[string]string

	err := r.db.QueryRow(ctx, query, text, language).Scan(
		&w.ID,
		&w.Text,
		&w.Language,
		&definitions,
		&examples,
		&w.Pronunciation,
		&w.Etymology,
		&translations,
		&w.CreatedAt,
		&w.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, word.ErrWordNotFound
		}
		return nil, fmt.Errorf("failed to query word: %w", err)
	}

	w.Definitions = definitions
	w.Examples = examples
	w.Translations = translations

	return &w, nil
}

// Save stores a word in the repository
func (r *WordRepository) Save(ctx context.Context, w *word.Word) error {
	query := `
		INSERT INTO words (text, language, definitions, examples, pronunciation, etymology, translations, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (text, language) 
		DO UPDATE SET 
			definitions = $3,
			examples = $4,
			pronunciation = $5,
			etymology = $6,
			translations = $7,
			updated_at = $9
		RETURNING id
	`

	now := time.Now()
	w.UpdatedAt = now
	if w.CreatedAt.IsZero() {
		w.CreatedAt = now
	}

	return r.db.QueryRow(ctx, query,
		w.Text,
		w.Language,
		w.Definitions,
		w.Examples,
		w.Pronunciation,
		w.Etymology,
		w.Translations,
		w.CreatedAt,
		w.UpdatedAt,
	).Scan(&w.ID)
}

// List retrieves words with optional filtering
func (r *WordRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*word.Word, error) {
	// Build query with filters
	query := `
		SELECT id, text, language, definitions, examples, pronunciation, etymology, translations, created_at, updated_at
		FROM words
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	// Add filters
	for key, value := range filter {
		query += fmt.Sprintf(" AND %s = $%d", key, argIndex)
		args = append(args, value)
		argIndex++
	}

	// Add ordering and pagination
	query += " ORDER BY updated_at DESC LIMIT $" + fmt.Sprintf("%d", argIndex)
	args = append(args, limit)
	argIndex++

	if offset > 0 {
		query += " OFFSET $" + fmt.Sprintf("%d", argIndex)
		args = append(args, offset)
	}

	// Execute query
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query words: %w", err)
	}
	defer rows.Close()

	// Process results
	var words []*word.Word
	for rows.Next() {
		var w word.Word
		var definitions, examples []string
		var translations map[string]string

		err := rows.Scan(
			&w.ID,
			&w.Text,
			&w.Language,
			&definitions,
			&examples,
			&w.Pronunciation,
			&w.Etymology,
			&translations,
			&w.CreatedAt,
			&w.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan word row: %w", err)
		}

		w.Definitions = definitions
		w.Examples = examples
		w.Translations = translations

		words = append(words, &w)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating word rows: %w", err)
	}

	return words, nil
}
