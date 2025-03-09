package repository

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
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
		db:     mock,
		logger: logger.With().Str("component", "word_repository_test").Logger(),
	}

	return mock, repo
}

// Basic error handling test - this is the only test we keep in the unit tests
// since it's easier to test with mocks
func TestWordRepository_ErrorHandling(t *testing.T) {
	mock, repo := setupMockDB(t)
	defer mock.Close()

	ctx := context.Background()

	// Test database error handling
	mock.ExpectQuery(`SELECT id, text, language, definitions, etymology, translations, 
		       search_terms, lemma, created_at, updated_at
		FROM words
		WHERE text = \$1 AND language = \$2`).
		WithArgs("error-test", "en").
		WillReturnError(pgx.ErrNoRows)

	// Execute the function being tested
	result, err := repo.FindByText(ctx, "error-test", "en")

	// Verify results
	assert.Nil(t, result)
	assert.ErrorIs(t, err, word.ErrWordNotFound)

	// Test with a different error
	mock.ExpectQuery(`SELECT id, text, language, definitions, etymology, translations, 
		       search_terms, lemma, created_at, updated_at
		FROM words
		WHERE text = \$1 AND language = \$2`).
		WithArgs("db-error", "en").
		WillReturnError(context.DeadlineExceeded)

	// Execute the function being tested
	result, err = repo.FindByText(ctx, "db-error", "en")

	// Verify results
	assert.Nil(t, result)
	assert.ErrorContains(t, err, "failed to query word")
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	// Ensure all expectations were met
	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}
