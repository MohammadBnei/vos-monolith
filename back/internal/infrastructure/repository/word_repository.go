package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

// Ensure WordRepository implements word.Repository
var _ word.Repository = (*WordRepository)(nil)

// NewWordRepository creates a new word repository
func NewWordRepository(db DBInterface, logger zerolog.Logger) *WordRepository {
	return &WordRepository{
		db:     db,
		logger: logger.With().Str("component", "word_repository").Logger(),
	}
}

// FindByID retrieves a word by its ID
func (r *WordRepository) FindByID(ctx context.Context, id string) (*word.Word, error) {
	r.logger.Debug().Str("id", id).Msg("Finding word by ID")

	query := `
		SELECT id, text, language, definitions, etymology, translations, 
		       synonyms, antonyms, search_terms, lemma, usage_notes, pronunciation, created_at, updated_at
		FROM words
		WHERE id = $1
	`

	var w word.Word
	var definitionsJSON []byte
	var searchTerms []string
	var translations map[string][]string
	var synonyms []string
	var antonyms []string
	var usageNotes []string
	var pronunciation string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&w.ID,
		&w.Text,
		&w.Language,
		&definitionsJSON,
		&w.Etymology,
		&translations,
		&synonyms,
		&antonyms,
		&searchTerms,
		&w.Lemma,
		&usageNotes,
		&pronunciation,
		&w.CreatedAt,
		&w.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, word.ErrWordNotFound
		}
		return nil, fmt.Errorf("failed to query word by ID: %w", err)
	}

	// Parse definitions JSON
	if err := json.Unmarshal(definitionsJSON, &w.Definitions); err != nil {
		return nil, fmt.Errorf("failed to parse definitions: %w", err)
	}

	w.SearchTerms = searchTerms
	w.Translations = translations
	w.Synonyms = synonyms
	w.Antonyms = antonyms
	w.UsageNotes = usageNotes
	w.Pronunciation = pronunciation
	w.Pronunciation = pronunciation
	w.Pronunciation = pronunciation

	return &w, nil
}

// FindByText retrieves a word by its text and language
func (r *WordRepository) FindByText(ctx context.Context, text, language string) (*word.Word, error) {
	query := `
		SELECT id, text, language, definitions, etymology, translations, 
		       synonyms, antonyms, search_terms, lemma, usage_notes, pronunciation, created_at, updated_at
		FROM words
		WHERE text = $1 AND language = $2
	`

	var w word.Word
	var definitionsJSON []byte
	var searchTerms []string
	var translations map[string][]string
	var synonyms []string
	var antonyms []string
	var usageNotes []string
	var pronunciation string

	err := r.db.QueryRow(ctx, query, text, language).Scan(
		&w.ID,
		&w.Text,
		&w.Language,
		&definitionsJSON,
		&w.Etymology,
		&translations,
		&synonyms,
		&antonyms,
		&searchTerms,
		&w.Lemma,
		&usageNotes,
		&pronunciation,
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

	w.SearchTerms = searchTerms
	w.Translations = translations
	w.Synonyms = synonyms
	w.Antonyms = antonyms
	w.UsageNotes = usageNotes

	return &w, nil
}

// FindByAnyForm retrieves a word by any of its forms (using search terms)
func (r *WordRepository) FindByAnyForm(ctx context.Context, text, language string) (*word.Word, error) {
	query := `
		SELECT id, text, language, definitions, etymology, translations, 
		        synonyms, antonyms, search_terms, lemma, usage_notes, pronunciation, created_at, updated_at
		FROM words
		WHERE language = $1 AND $2 = ANY(search_terms)
	`

	var w word.Word
	var definitionsJSON []byte
	var searchTerms []string
	var translations map[string][]string
	var synonyms []string
	var antonyms []string
	var usageNotes []string
	var pronunciation string

	err := r.db.QueryRow(ctx, query, language, text).Scan(
		&w.ID,
		&w.Text,
		&w.Language,
		&definitionsJSON,
		&w.Etymology,
		&translations,
		&synonyms,
		&antonyms,
		&searchTerms,
		&w.Lemma,
		&usageNotes,
		&pronunciation,
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

	w.SearchTerms = searchTerms
	w.Translations = translations
	w.Synonyms = synonyms
	w.Antonyms = antonyms
	w.UsageNotes = usageNotes

	return &w, nil
}

// Save stores a word in the repository
func (r *WordRepository) Save(ctx context.Context, w *word.Word) error {
	query := `
		INSERT INTO words (
			id, text, language, definitions, etymology, translations, 
			synonyms, antonyms, search_terms, lemma, usage_notes, pronunciation, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (text, language) 
		DO UPDATE SET 
			definitions = $4,
			etymology = $5,
			translations = $6,
			synonyms = $7,
			antonyms = $8,
			search_terms = $9,
			lemma = $10,
			usage_notes = $11,
			pronunciation = $12,
			created_at = $13,
			updated_at = $14
		RETURNING id
	`

	now := time.Now()
	w.UpdatedAt = now
	if w.CreatedAt.IsZero() {
		w.CreatedAt = now
	}

	// Generate ID if not set
	if w.ID == "" {
		w.ID = uuid.New().String()
	}

	// Convert definitions to JSON
	definitionsJSON, err := json.Marshal(w.Definitions)
	if err != nil {
		return fmt.Errorf("failed to marshal definitions: %w", err)
	}

	return r.db.QueryRow(ctx, query,
		w.ID,
		w.Text,
		w.Language,
		definitionsJSON,
		w.Etymology,
		w.Translations,
		w.Synonyms,
		w.Antonyms,
		w.SearchTerms,
		w.Lemma,
		w.UsageNotes,
		w.Pronunciation,
		w.CreatedAt,
		w.UpdatedAt,
	).Scan(&w.ID)
}

// FindByPrefix retrieves words by prefix and language
func (r *WordRepository) FindByPrefix(ctx context.Context, prefix, language string, limit int) ([]*word.Word, error) {
	query := `
		SELECT id, text, language, definitions, etymology, 
		       translations, synonyms, antonyms, search_terms, lemma, usage_notes, pronunciation, created_at, updated_at
		FROM words
		WHERE language = $1 
		  AND EXISTS (SELECT 1 FROM unnest(search_terms) AS term 
		         WHERE term ILIKE $2 || '%')
		ORDER BY text <-> $2  -- Use pg_trgm similarity operator
		LIMIT $3
	`

	var words []*word.Word
	rows, err := r.db.Query(ctx, query, language, prefix, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query words by prefix: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var w word.Word
		var definitionsJSON []byte
		var searchTerms []string
		var translations map[string][]string
		var synonyms []string
		var antonyms []string
		var usageNotes []string
		var pronunciation string

		if err := rows.Scan(
			&w.ID,
			&w.Text,
			&w.Language,
			&definitionsJSON,
			&w.Etymology,
			&translations,
			&synonyms,
			&antonyms,
			&searchTerms,
			&w.Lemma,
			&usageNotes,
			&pronunciation,
			&w.CreatedAt,
			&w.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan word row: %w", err)
		}

		if err := json.Unmarshal(definitionsJSON, &w.Definitions); err != nil {
			return nil, fmt.Errorf("failed to parse definitions: %w", err)
		}

		w.SearchTerms = searchTerms
		w.Translations = translations
		w.Pronunciation = pronunciation
		w.Synonyms = synonyms
		w.Antonyms = antonyms
		w.UsageNotes = usageNotes
		w.Pronunciation = pronunciation

		words = append(words, &w)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating word rows: %w", err)
	}

	return words, nil
}

func (r *WordRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*word.Word, error) {
	// Build query with filters
	query := `
		SELECT id, text, language, definitions, etymology, translations, 
		        search_terms, lemma, pronunciation, created_at, updated_at
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
		var definitionsJSON []byte
		var searchTerms []string
		var translations map[string][]string
		var pronunciation string

		err := rows.Scan(
			&w.ID,
			&w.Text,
			&w.Language,
			&definitionsJSON,
			&w.Etymology,
			&translations,
			&searchTerms,
			&w.Lemma,
			&pronunciation,
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

		w.SearchTerms = searchTerms
		w.Translations = translations

		words = append(words, &w)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating word rows: %w", err)
	}

	return words, nil
}

// FindSuggestions retrieves word suggestions based on a prefix
func (r *WordRepository) FindSuggestions(ctx context.Context, prefix, language string, limit int) ([]string, error) {
	r.logger.Debug().Str("prefix", prefix).Str("language", language).Int("limit", limit).Msg("Finding suggestions")

	query := `
		SELECT text
		FROM words
		WHERE language = $1 AND text ILIKE $2 || '%'
		ORDER BY updated_at DESC
		LIMIT $3
	`

	rows, err := r.db.Query(ctx, query, language, prefix, limit)
	if err != nil {
		r.logger.Error().Err(err).Msg("Error finding suggestions")
		return nil, fmt.Errorf("error finding suggestions: %w", err)
	}
	defer rows.Close()

	var suggestions []string
	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			r.logger.Error().Err(err).Msg("Error scanning suggestion")
			continue
		}
		suggestions = append(suggestions, text)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error().Err(err).Msg("Error iterating suggestions")
		return nil, fmt.Errorf("error iterating suggestions: %w", err)
	}

	return suggestions, nil
}
