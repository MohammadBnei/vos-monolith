package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"voconsteroid/internal/domain/word"
)

// DBInterface defines the database operations needed by WordRepository
type DBInterface interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

// WordRepository implements the word.Repository interface using PostgreSQL
type WordRepository struct {
	db     DBInterface
	logger zerolog.Logger
}

// NewWordRepository creates a new word repository
func NewWordRepository(db DBInterface, logger zerolog.Logger) *WordRepository {
	return &WordRepository{
		db:     db,
		logger: logger.With().Str("component", "word_repository").Logger(),
	}
}

// FindByText retrieves a word by its text and language
func (r *WordRepository) FindByText(ctx context.Context, text, language string) (*word.Word, error) {
	query := `
		SELECT id, text, language, definitions, examples, pronunciation, etymology, translations, 
		       word_type, forms, search_terms, lemma, created_at, updated_at
		FROM words
		WHERE text = $1 AND language = $2
	`

	var w word.Word
	var definitionsJSON, wordFormsJSON []byte
	var examples, searchTerms []string
	var translations map[string]string

	err := r.db.QueryRow(ctx, query, text, language).Scan(
		&w.ID,
		&w.Text,
		&w.Language,
		&definitionsJSON,
		&examples,
		&w.Pronunciation,
		&w.Etymology,
		&translations,
		&w.WordType,
		&wordFormsJSON,
		&searchTerms,
		&w.Lemma,
		&w.CreatedAt,
		&w.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, word.ErrWordNotFound
		}
		return nil, fmt.Errorf("failed to query word: %w", err)
	}

	// Parse definitions JSON
	if err := json.Unmarshal(definitionsJSON, &w.Definitions); err != nil {
		return nil, fmt.Errorf("failed to parse definitions: %w", err)
	}

	// Parse word forms JSON
	if err := json.Unmarshal(wordFormsJSON, &w.Forms); err != nil {
		return nil, fmt.Errorf("failed to parse word forms: %w", err)
	}

	w.Examples = examples
	w.SearchTerms = searchTerms
	w.Translations = translations

	return &w, nil
}

// FindByAnyForm retrieves a word by any of its forms (using search terms)
func (r *WordRepository) FindByAnyForm(ctx context.Context, text, language string) (*word.Word, error) {
	query := `
		SELECT id, text, language, definitions, examples, pronunciation, etymology, translations, 
		       word_type, forms, search_terms, lemma, created_at, updated_at
		FROM words
		WHERE language = $1 AND $2 = ANY(search_terms)
	`

	var w word.Word
	var definitionsJSON, wordFormsJSON []byte
	var examples, searchTerms []string
	var translations map[string]string

	err := r.db.QueryRow(ctx, query, language, text).Scan(
		&w.ID,
		&w.Text,
		&w.Language,
		&definitionsJSON,
		&examples,
		&w.Pronunciation,
		&w.Etymology,
		&translations,
		&w.WordType,
		&wordFormsJSON,
		&searchTerms,
		&w.Lemma,
		&w.CreatedAt,
		&w.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, word.ErrWordNotFound
		}
		return nil, fmt.Errorf("failed to query word by form: %w", err)
	}

	// Parse definitions JSON
	if err := json.Unmarshal(definitionsJSON, &w.Definitions); err != nil {
		return nil, fmt.Errorf("failed to parse definitions: %w", err)
	}

	// Parse word forms JSON
	if err := json.Unmarshal(wordFormsJSON, &w.Forms); err != nil {
		return nil, fmt.Errorf("failed to parse word forms: %w", err)
	}

	w.Examples = examples
	w.SearchTerms = searchTerms
	w.Translations = translations

	return &w, nil
}

// Save stores a word in the repository
func (r *WordRepository) Save(ctx context.Context, w *word.Word) error {
	query := `
		INSERT INTO words (
			text, language, definitions, examples, pronunciation, etymology, translations, 
			word_type, forms, search_terms, lemma, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (text, language) 
		DO UPDATE SET 
			definitions = $3,
			examples = $4,
			pronunciation = $5,
			etymology = $6,
			translations = $7,
			word_type = $8,
			forms = $9,
			search_terms = $10,
			lemma = $11,
			updated_at = $13
		RETURNING id
	`

	now := time.Now()
	w.UpdatedAt = now
	if w.CreatedAt.IsZero() {
		w.CreatedAt = now
	}

	// Convert definitions to JSON
	definitionsJSON, err := json.Marshal(w.Definitions)
	if err != nil {
		return fmt.Errorf("failed to marshal definitions: %w", err)
	}

	// Convert word forms to JSON
	wordFormsJSON, err := json.Marshal(w.Forms)
	if err != nil {
		return fmt.Errorf("failed to marshal word forms: %w", err)
	}

	return r.db.QueryRow(ctx, query,
		w.Text,
		w.Language,
		definitionsJSON,
		w.Examples,
		w.Pronunciation,
		w.Etymology,
		w.Translations,
		w.WordType,
		wordFormsJSON,
		w.SearchTerms,
		w.Lemma,
		w.CreatedAt,
		w.UpdatedAt,
	).Scan(&w.ID)
}

// List retrieves words with optional filtering
func (r *WordRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*word.Word, error) {
	// Build query with filters
	query := `
		SELECT id, text, language, definitions, examples, pronunciation, etymology, translations, 
		       word_type, forms, search_terms, lemma, created_at, updated_at
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
		var definitionsJSON, wordFormsJSON []byte
		var examples, searchTerms []string
		var translations map[string]string

		err := rows.Scan(
			&w.ID,
			&w.Text,
			&w.Language,
			&definitionsJSON,
			&examples,
			&w.Pronunciation,
			&w.Etymology,
			&translations,
			&w.WordType,
			&wordFormsJSON,
			&searchTerms,
			&w.Lemma,
			&w.CreatedAt,
			&w.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan word row: %w", err)
		}

		// Parse definitions JSON
		if err := json.Unmarshal(definitionsJSON, &w.Definitions); err != nil {
			return nil, fmt.Errorf("failed to parse definitions: %w", err)
		}

		// Parse word forms JSON
		if err := json.Unmarshal(wordFormsJSON, &w.Forms); err != nil {
			return nil, fmt.Errorf("failed to parse word forms: %w", err)
		}

		w.Examples = examples
		w.SearchTerms = searchTerms
		w.Translations = translations

		words = append(words, &w)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating word rows: %w", err)
	}

	return words, nil
}
