package repository_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"voconsteroid/internal/domain/word"
	"voconsteroid/internal/infrastructure/repository"
)

// setupTestDatabase creates a PostgreSQL container and returns a connection pool
func setupTestDatabase(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	ctx := context.Background()

	// Define PostgreSQL container using the postgres module
	dbName := "testdb"
	dbUser := "testuser"
	dbPassword := "testpass"

	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err)

	// Get connection string
	connString, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Connect to database
	dbpool, err := pgxpool.New(ctx, connString)
	require.NoError(t, err)

	// Run migrations instead of creating schema manually
	err = runMigrations(ctx, dbpool)
	require.NoError(t, err)

	// Return cleanup function
	cleanup := func() {
		dbpool.Close()
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}
	}

	return dbpool, cleanup
}

// runMigrations executes all SQL migration scripts against the database
func runMigrations(ctx context.Context, db *pgxpool.Pool) error {
	// Path to migration files
	migrationsPath := "../migrations"

	// Get absolute path to migrations directory
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path to migrations: %w", err)
	}

	// Read all SQL files from the migrations directory
	entries, err := filepath.Glob(filepath.Join(absPath, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	// Sort migration files to ensure they're executed in the correct order
	// This assumes migration files are named with a numeric prefix like "001_create_tables.sql"
	sort.Strings(entries)

	// Execute each migration file
	for _, entry := range entries {
		content, err := os.ReadFile(entry)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", entry, err)
		}

		_, err = db.Exec(ctx, string(content))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", entry, err)
		}
	}

	return nil
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
		Definitions: []word.Definition{
			{
				WordType: "noun",
			},
		},
		SearchTerms:  []string{"test", "tests"},
		Lemma:        "test",
		CreatedAt:    now,
		UpdatedAt:    now,
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

	t.Run("UpdateExistingWord", func(t *testing.T) {
		ctx := context.Background()

		// Create and save initial word
		testWord := createTestWord()
		testWord.Text = "update-test"

		err := repo.Save(ctx, testWord)
		require.NoError(t, err)
		assert.NotEmpty(t, testWord.ID)

		originalID := testWord.ID

		// Modify the word
		testWord.Definitions = append(testWord.Definitions, word.Definition{
			Text:     "a second definition",
			WordType: "verb",
			Examples: []string{"to test something"},
		})
		testWord.Examples = append(testWord.Examples, "another example")

		// Save the updated word
		err = repo.Save(ctx, testWord)
		require.NoError(t, err)

		// ID should remain the same
		assert.Equal(t, originalID, testWord.ID)

		// Retrieve the word and verify updates
		updated, err := repo.FindByText(ctx, testWord.Text, testWord.Language)
		require.NoError(t, err)

		assert.Equal(t, originalID, updated.ID)
		assert.Len(t, updated.Definitions, 2)
		assert.Equal(t, "a second definition", updated.Definitions[1].Text)
		assert.Contains(t, updated.Examples, "another example")
	})

	t.Run("ComplexFiltering", func(t *testing.T) {
		ctx := context.Background()

		// Create words with different languages
		enWord := createTestWord()
		enWord.Text = "filter-test-en"
		enWord.Language = "en"

		frWord := createTestWord()
		frWord.Text = "filter-test-fr"
		frWord.Language = "fr"

		esWord := createTestWord()
		esWord.Text = "filter-test-es"
		esWord.Language = "es"

		// Save all words
		require.NoError(t, repo.Save(ctx, enWord))
		require.NoError(t, repo.Save(ctx, frWord))
		require.NoError(t, repo.Save(ctx, esWord))

		// Filter by language
		filter := map[string]interface{}{"language": "fr"}
		results, err := repo.List(ctx, filter, 10, 0)
		require.NoError(t, err)

		// Should find at least the French word
		found := false
		for _, w := range results {
			if w.Text == frWord.Text {
				found = true
				assert.Equal(t, "fr", w.Language)
				break
			}
		}
		assert.True(t, found, "French word should be found in results")

		// Should not find English or Spanish words in French results
		for _, w := range results {
			assert.NotEqual(t, enWord.Text, w.Text, "English word should not be in French results")
			assert.NotEqual(t, esWord.Text, w.Text, "Spanish word should not be in French results")
		}
	})

	t.Run("ConcurrentOperations", func(t *testing.T) {
		// Test concurrent operations on the repository
		const numGoroutines = 5
		const wordsPerGoroutine = 3

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(routineID int) {
				defer wg.Done()

				for j := 0; j < wordsPerGoroutine; j++ {
					ctx := context.Background()

					// Create a unique word for this goroutine
					testWord := createTestWord()
					testWord.Text = fmt.Sprintf("concurrent-test-%d-%d", routineID, j)

					// Save the word
					err := repo.Save(ctx, testWord)
					assert.NoError(t, err)

					// Find the word
					found, err := repo.FindByText(ctx, testWord.Text, testWord.Language)
					assert.NoError(t, err)
					assert.Equal(t, testWord.Text, found.Text)
				}
			}(i)
		}

		wg.Wait()

		// Verify we can find all the words
		ctx := context.Background()
		filter := map[string]interface{}{"language": "en"}
		results, err := repo.List(ctx, filter, 100, 0)
		require.NoError(t, err)

		// Count concurrent test words
		concurrentWords := 0
		for _, w := range results {
			if strings.HasPrefix(w.Text, "concurrent-test-") {
				concurrentWords++
			}
		}

		assert.Equal(t, numGoroutines*wordsPerGoroutine, concurrentWords,
			"Should find all concurrently created words")
	})
}
