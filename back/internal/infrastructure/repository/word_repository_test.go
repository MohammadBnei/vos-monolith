package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"voconsteroid/internal/domain/word"
)

// setupMockDB sets up a mock database for testing
func setupMockDB(t *testing.T) (pgxmock.PgxPoolIface, *WordRepository) {
	t.Helper()

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)

	logger := zerolog.New(zerolog.NewTestWriter(t))
	repo := &WordRepository{
		db:     mock.(interface{}).(*pgxpool.Pool),
		logger: logger.With().Str("component", "word_repository_test").Logger(),
	}

	return mock, repo
}

// createTestWord creates a test word for use in tests
func createTestWord() *word.Word {
	now := time.Now()
	return &word.Word{
		ID:       "test-id",
		Text:     "test",
		Language: "en",
		Definitions: []word.Definition{
			{
				Text:     "a procedure intended to establish the quality, performance, or reliability of something",
				WordType: "noun",
				Examples: []string{"the car has passed its test"},
			},
		},
		Examples: []string{
			"we need to test this code",
			"testing is important",
		},
		Pronunciation: map[string]string{
			"IPA": "/tɛst/",
		},
		Etymology:    "from Latin testum, meaning 'earthen pot'",
		Translations: map[string]string{"fr": "test", "es": "prueba"},
		WordType:     "noun",
		Forms: []word.Form{
			{
				Text:       "test",
				Attributes: map[string]string{"number": "singular"},
				IsLemma:    true,
			},
			{
				Text:       "tests",
				Attributes: map[string]string{"number": "plural"},
				IsLemma:    false,
			},
		},
		SearchTerms: []string{"test", "tests"},
		Lemma:       "test",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestWordRepository_FindByText(t *testing.T) {
	mock, repo := setupMockDB(t)
	defer mock.Close()

	testWord := createTestWord()
	ctx := context.Background()

	// Setup expected query and response
	rows := pgxmock.NewRows([]string{
		"id", "text", "language", "definitions", "examples", "pronunciation",
		"etymology", "translations", "word_type", "forms", "search_terms",
		"lemma", "created_at", "updated_at",
	}).AddRow(
		testWord.ID,
		testWord.Text,
		testWord.Language,
		`[{"text":"a procedure intended to establish the quality, performance, or reliability of something","word_type":"noun","examples":["the car has passed its test"]}]`,
		testWord.Examples,
		testWord.Pronunciation,
		testWord.Etymology,
		testWord.Translations,
		testWord.WordType,
		`[{"text":"test","attributes":{"number":"singular"},"is_lemma":true},{"text":"tests","attributes":{"number":"plural"},"is_lemma":false}]`,
		testWord.SearchTerms,
		testWord.Lemma,
		testWord.CreatedAt,
		testWord.UpdatedAt,
	)

	mock.ExpectQuery(`SELECT id, text, language, definitions, examples, pronunciation, etymology, translations, 
		       word_type, forms, search_terms, lemma, created_at, updated_at
		FROM words
		WHERE text = \$1 AND language = \$2`).
		WithArgs(testWord.Text, testWord.Language).
		WillReturnRows(rows)

	// Execute the function being tested
	result, err := repo.FindByText(ctx, testWord.Text, testWord.Language)

	// Verify results
	require.NoError(t, err)
	assert.Equal(t, testWord.ID, result.ID)
	assert.Equal(t, testWord.Text, result.Text)
	assert.Equal(t, testWord.Language, result.Language)
	assert.Equal(t, testWord.WordType, result.WordType)
	assert.Equal(t, testWord.Lemma, result.Lemma)
	assert.Len(t, result.Definitions, 1)
	assert.Equal(t, testWord.Definitions[0].Text, result.Definitions[0].Text)
	assert.Len(t, result.Forms, 2)
	assert.Equal(t, testWord.Forms[0].Text, result.Forms[0].Text)
	assert.Equal(t, testWord.Forms[1].Text, result.Forms[1].Text)

	// Ensure all expectations were met
	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestWordRepository_FindByText_NotFound(t *testing.T) {
	mock, repo := setupMockDB(t)
	defer mock.Close()

	ctx := context.Background()

	// Setup expected query with no results
	mock.ExpectQuery(`SELECT id, text, language, definitions, examples, pronunciation, etymology, translations, 
		       word_type, forms, search_terms, lemma, created_at, updated_at
		FROM words
		WHERE text = \$1 AND language = \$2`).
		WithArgs("nonexistent", "en").
		WillReturnError(errors.New("no rows in result set"))

	// Execute the function being tested
	result, err := repo.FindByText(ctx, "nonexistent", "en")

	// Verify results
	assert.Nil(t, result)
	assert.ErrorIs(t, err, word.ErrWordNotFound)

	// Ensure all expectations were met
	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestWordRepository_FindByAnyForm(t *testing.T) {
	mock, repo := setupMockDB(t)
	defer mock.Close()

	testWord := createTestWord()
	ctx := context.Background()

	// Setup expected query and response
	rows := pgxmock.NewRows([]string{
		"id", "text", "language", "definitions", "examples", "pronunciation",
		"etymology", "translations", "word_type", "forms", "search_terms",
		"lemma", "created_at", "updated_at",
	}).AddRow(
		testWord.ID,
		testWord.Text,
		testWord.Language,
		`[{"text":"a procedure intended to establish the quality, performance, or reliability of something","word_type":"noun","examples":["the car has passed its test"]}]`,
		testWord.Examples,
		testWord.Pronunciation,
		testWord.Etymology,
		testWord.Translations,
		testWord.WordType,
		`[{"text":"test","attributes":{"number":"singular"},"is_lemma":true},{"text":"tests","attributes":{"number":"plural"},"is_lemma":false}]`,
		testWord.SearchTerms,
		testWord.Lemma,
		testWord.CreatedAt,
		testWord.UpdatedAt,
	)

	mock.ExpectQuery(`SELECT id, text, language, definitions, examples, pronunciation, etymology, translations, 
		       word_type, forms, search_terms, lemma, created_at, updated_at
		FROM words
		WHERE language = \$1 AND \$2 = ANY\(search_terms\)`).
		WithArgs(testWord.Language, "tests").
		WillReturnRows(rows)

	// Execute the function being tested
	result, err := repo.FindByAnyForm(ctx, "tests", testWord.Language)

	// Verify results
	require.NoError(t, err)
	assert.Equal(t, testWord.ID, result.ID)
	assert.Equal(t, testWord.Text, result.Text)
	assert.Equal(t, testWord.Language, result.Language)
	assert.Len(t, result.Forms, 2)
	assert.Equal(t, "tests", result.Forms[1].Text)

	// Ensure all expectations were met
	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestWordRepository_Save(t *testing.T) {
	mock, repo := setupMockDB(t)
	defer mock.Close()

	testWord := createTestWord()
	ctx := context.Background()

	// Setup expected query and response
	rows := pgxmock.NewRows([]string{"id"}).AddRow(testWord.ID)

	mock.ExpectQuery(`INSERT INTO words \(
			text, language, definitions, examples, pronunciation, etymology, translations, 
			word_type, forms, search_terms, lemma, created_at, updated_at
		\)
		VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8, \$9, \$10, \$11, \$12, \$13\)
		ON CONFLICT \(text, language\) 
		DO UPDATE SET 
			definitions = \$3,
			examples = \$4,
			pronunciation = \$5,
			etymology = \$6,
			translations = \$7,
			word_type = \$8,
			forms = \$9,
			search_terms = \$10,
			lemma = \$11,
			updated_at = \$13
		RETURNING id`).
		WithArgs(
			testWord.Text,
			testWord.Language,
			pgxmock.AnyArg(), // definitions JSON
			testWord.Examples,
			testWord.Pronunciation,
			testWord.Etymology,
			testWord.Translations,
			testWord.WordType,
			pgxmock.AnyArg(), // forms JSON
			testWord.SearchTerms,
			testWord.Lemma,
			pgxmock.AnyArg(), // created_at
			pgxmock.AnyArg(), // updated_at
		).
		WillReturnRows(rows)

	// Execute the function being tested
	err := repo.Save(ctx, testWord)

	// Verify results
	require.NoError(t, err)
	assert.Equal(t, testWord.ID, testWord.ID)

	// Ensure all expectations were met
	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestWordRepository_List(t *testing.T) {
	mock, repo := setupMockDB(t)
	defer mock.Close()

	testWord := createTestWord()
	ctx := context.Background()
	filter := map[string]interface{}{"language": "en"}
	limit := 10
	offset := 0

	// Setup expected query and response
	rows := pgxmock.NewRows([]string{
		"id", "text", "language", "definitions", "examples", "pronunciation",
		"etymology", "translations", "word_type", "forms", "search_terms",
		"lemma", "created_at", "updated_at",
	}).AddRow(
		testWord.ID,
		testWord.Text,
		testWord.Language,
		`[{"text":"a procedure intended to establish the quality, performance, or reliability of something","word_type":"noun","examples":["the car has passed its test"]}]`,
		testWord.Examples,
		testWord.Pronunciation,
		testWord.Etymology,
		testWord.Translations,
		testWord.WordType,
		`[{"text":"test","attributes":{"number":"singular"},"is_lemma":true},{"text":"tests","attributes":{"number":"plural"},"is_lemma":false}]`,
		testWord.SearchTerms,
		testWord.Lemma,
		testWord.CreatedAt,
		testWord.UpdatedAt,
	)

	mock.ExpectQuery(`SELECT id, text, language, definitions, examples, pronunciation, etymology, translations, 
		       word_type, forms, search_terms, lemma, created_at, updated_at
		FROM words
		WHERE 1=1 AND language = \$1 ORDER BY updated_at DESC LIMIT \$2`).
		WithArgs("en", limit).
		WillReturnRows(rows)

	// Execute the function being tested
	results, err := repo.List(ctx, filter, limit, offset)

	// Verify results
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, testWord.ID, results[0].ID)
	assert.Equal(t, testWord.Text, results[0].Text)
	assert.Equal(t, testWord.Language, results[0].Language)

	// Ensure all expectations were met
	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestWordRepository_List_WithOffset(t *testing.T) {
	mock, repo := setupMockDB(t)
	defer mock.Close()

	testWord := createTestWord()
	ctx := context.Background()
	filter := map[string]interface{}{"language": "en"}
	limit := 10
	offset := 5

	// Setup expected query and response
	rows := pgxmock.NewRows([]string{
		"id", "text", "language", "definitions", "examples", "pronunciation",
		"etymology", "translations", "word_type", "forms", "search_terms",
		"lemma", "created_at", "updated_at",
	}).AddRow(
		testWord.ID,
		testWord.Text,
		testWord.Language,
		`[{"text":"a procedure intended to establish the quality, performance, or reliability of something","word_type":"noun","examples":["the car has passed its test"]}]`,
		testWord.Examples,
		testWord.Pronunciation,
		testWord.Etymology,
		testWord.Translations,
		testWord.WordType,
		`[{"text":"test","attributes":{"number":"singular"},"is_lemma":true},{"text":"tests","attributes":{"number":"plural"},"is_lemma":false}]`,
		testWord.SearchTerms,
		testWord.Lemma,
		testWord.CreatedAt,
		testWord.UpdatedAt,
	)

	mock.ExpectQuery(`SELECT id, text, language, definitions, examples, pronunciation, etymology, translations, 
		       word_type, forms, search_terms, lemma, created_at, updated_at
		FROM words
		WHERE 1=1 AND language = \$1 ORDER BY updated_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs("en", limit, offset).
		WillReturnRows(rows)

	// Execute the function being tested
	results, err := repo.List(ctx, filter, limit, offset)

	// Verify results
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, testWord.ID, results[0].ID)

	// Ensure all expectations were met
	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}
