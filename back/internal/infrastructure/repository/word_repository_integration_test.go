package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"voconsteroid/internal/domain/word"
	"voconsteroid/internal/infrastructure/repository"
)

// setupTestDatabase creates a PostgreSQL container and returns a connection pool
func setupTestDatabase(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	ctx := context.Background()

	// Define PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	// Start container
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Get host and port
	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Connect to database
	connString := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb", host, port.Port())
	dbpool, err := pgxpool.New(ctx, connString)
	require.NoError(t, err)

	// Create schema
	_, err = dbpool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS words (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			text TEXT NOT NULL,
			language TEXT NOT NULL,
			definitions JSONB NOT NULL DEFAULT '[]',
			examples TEXT[] NOT NULL DEFAULT '{}',
			pronunciation JSONB NOT NULL DEFAULT '{}',
			etymology TEXT,
			translations JSONB NOT NULL DEFAULT '{}',
			word_type TEXT,
			forms JSONB NOT NULL DEFAULT '[]',
			search_terms TEXT[] NOT NULL DEFAULT '{}',
			lemma TEXT,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
			UNIQUE(text, language)
		)
	`)
	require.NoError(t, err)

	// Return cleanup function
	cleanup := func() {
		dbpool.Close()
		container.Terminate(ctx)
	}

	return dbpool, cleanup
}

// createTestWord creates a test word for use in tests
func createTestWord() *word.Word {
	now := time.Now()
	return &word.Word{
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

func TestWordRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database
	dbpool, cleanup := setupTestDatabase(t)
	defer cleanup()

	// Create repository
	logger := zerolog.New(zerolog.NewTestWriter(t))
	repo := repository.NewWordRepository(dbpool, logger)

	// Run tests with real database
	t.Run("SaveAndFindByText", func(t *testing.T) {
		ctx := context.Background()

		// Create test word
		testWord := createTestWord()

		// Save word
		err := repo.Save(ctx, testWord)
		require.NoError(t, err)
		assert.NotEmpty(t, testWord.ID)

		// Find word
		found, err := repo.FindByText(ctx, testWord.Text, testWord.Language)
		require.NoError(t, err)
		assert.Equal(t, testWord.ID, found.ID)
		assert.Equal(t, testWord.Text, found.Text)
		assert.Equal(t, testWord.Language, found.Language)
		assert.Len(t, found.Definitions, 1)
		assert.Equal(t, testWord.Definitions[0].Text, found.Definitions[0].Text)
	})

	t.Run("FindByAnyForm", func(t *testing.T) {
		ctx := context.Background()

		// Create test word
		testWord := createTestWord()

		// Save word
		err := repo.Save(ctx, testWord)
		require.NoError(t, err)

		// Find word by form
		found, err := repo.FindByAnyForm(ctx, "tests", testWord.Language)
		require.NoError(t, err)
		assert.Equal(t, testWord.ID, found.ID)
		assert.Equal(t, testWord.Text, found.Text)
	})

	t.Run("List", func(t *testing.T) {
		ctx := context.Background()

		// Create test words
		testWord1 := createTestWord()
		testWord1.Text = "apple"
		testWord1.SearchTerms = []string{"apple", "apples"}

		testWord2 := createTestWord()
		testWord2.Text = "banana"
		testWord2.SearchTerms = []string{"banana", "bananas"}

		// Save words
		err := repo.Save(ctx, testWord1)
		require.NoError(t, err)

		err = repo.Save(ctx, testWord2)
		require.NoError(t, err)

		// List words
		filter := map[string]interface{}{"language": "en"}
		words, err := repo.List(ctx, filter, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(words), 2)

		// Test with limit
		limitedWords, err := repo.List(ctx, filter, 1, 0)
		require.NoError(t, err)
		assert.Len(t, limitedWords, 1)

		// Test with offset
		offsetWords, err := repo.List(ctx, filter, 10, 1)
		require.NoError(t, err)
		assert.Len(t, offsetWords, len(words)-1)
	})

	t.Run("NotFound", func(t *testing.T) {
		ctx := context.Background()

		// Try to find non-existent word
		_, err := repo.FindByText(ctx, "nonexistent", "en")
		assert.ErrorIs(t, err, word.ErrWordNotFound)

		// Try to find by non-existent form
		_, err = repo.FindByAnyForm(ctx, "nonexistentform", "en")
		assert.ErrorIs(t, err, word.ErrWordNotFound)
	})
}
