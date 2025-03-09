package word

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// RepositoryTestSuite defines a set of tests that any implementation of Repository should pass
type RepositoryTestSuite struct {
	Repo Repository
}

// TestFindByText tests the FindByText method of a Repository implementation
func (s *RepositoryTestSuite) TestFindByText(t *testing.T) {
	// Setup
	ctx := context.Background()
	testWord := &Word{
		Text:     "test",
		Language: "en",
		Definitions: []Definition{
			{
				Text:     "a procedure intended to establish the quality, performance, or reliability of something",
				WordType: "noun",
			},
		},
	}

	// Save the word first
	err := s.Repo.Save(ctx, testWord)
	assert.NoError(t, err)

	// Execute
	foundWord, err := s.Repo.FindByText(ctx, "test", "en")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, foundWord)
	assert.Equal(t, testWord.Text, foundWord.Text)
	assert.Equal(t, testWord.Language, foundWord.Language)
	assert.Equal(t, testWord.Definitions[0].Text, foundWord.Definitions[0].Text)
	assert.Equal(t, testWord.Definitions[0].WordType, foundWord.Definitions[0].WordType)
}

// TestFindByText_NotFound tests the FindByText method when the word doesn't exist
func (s *RepositoryTestSuite) TestFindByText_NotFound(t *testing.T) {
	// Setup
	ctx := context.Background()

	// Execute
	foundWord, err := s.Repo.FindByText(ctx, "nonexistent", "en")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, foundWord)
}

// TestSave tests the Save method of a Repository implementation
func (s *RepositoryTestSuite) TestSave(t *testing.T) {
	// Setup
	ctx := context.Background()
	testWord := &Word{
		Text:     "save_test",
		Language: "en",
		Definitions: []Definition{
			{
				Text:     "to keep safe or rescue from harm or danger",
				WordType: "verb",
			},
		},
	}

	// Execute
	err := s.Repo.Save(ctx, testWord)

	// Assert
	assert.NoError(t, err)

	// Verify the word was saved by retrieving it
	foundWord, err := s.Repo.FindByText(ctx, "save_test", "en")
	assert.NoError(t, err)
	assert.NotNil(t, foundWord)
	assert.Equal(t, testWord.Text, foundWord.Text)
}

// TestSave_Update tests updating an existing word
func (s *RepositoryTestSuite) TestSave_Update(t *testing.T) {
	// Setup
	ctx := context.Background()
	testWord := &Word{
		Text:     "update_test",
		Language: "en",
		Definitions: []Definition{
			{
				Text:     "original definition",
				WordType: "noun",
			},
		},
	}

	// Save the word first
	err := s.Repo.Save(ctx, testWord)
	assert.NoError(t, err)

	// Update the word
	testWord.Definitions = []Definition{
		{
			Text:     "updated definition",
			WordType: "noun",
		},
	}
	err = s.Repo.Save(ctx, testWord)
	assert.NoError(t, err)

	// Verify the word was updated
	foundWord, err := s.Repo.FindByText(ctx, "update_test", "en")
	assert.NoError(t, err)
	assert.NotNil(t, foundWord)
	assert.Equal(t, "updated definition", foundWord.Definitions[0].Text)
}

// TestList tests the List method of a Repository implementation
func (s *RepositoryTestSuite) TestList(t *testing.T) {
	// Setup
	ctx := context.Background()
	
	// Save some test words
	testWords := []*Word{
		{Text: "list_test1", Language: "en", Definitions: []Definition{{Text: "test 1", WordType: "noun"}}},
		{Text: "list_test2", Language: "en", Definitions: []Definition{{Text: "test 2", WordType: "noun"}}},
		{Text: "list_test3", Language: "fr", Definitions: []Definition{{Text: "test 3", WordType: "noun"}}},
	}
	
	for _, w := range testWords {
		err := s.Repo.Save(ctx, w)
		assert.NoError(t, err)
	}

	// Execute - list English words
	filter := map[string]interface{}{"language": "en"}
	words, err := s.Repo.List(ctx, filter, 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(words), 2) // At least our 2 English test words
	
	// Check if our test words are in the results
	var foundTest1, foundTest2 bool
	for _, w := range words {
		if w.Text == "list_test1" && w.Language == "en" {
			foundTest1 = true
		}
		if w.Text == "list_test2" && w.Language == "en" {
			foundTest2 = true
		}
		// Should not find French word
		assert.NotEqual(t, "list_test3", w.Text)
	}
	
	assert.True(t, foundTest1, "list_test1 should be in the results")
	assert.True(t, foundTest2, "list_test2 should be in the results")
}

// TestList_Limit tests the limit parameter of the List method
func (s *RepositoryTestSuite) TestList_Limit(t *testing.T) {
	// Setup
	ctx := context.Background()
	
	// Save some test words
	for i := 0; i < 5; i++ {
		testWord := &Word{
			Text:     "limit_test" + string(rune('a'+i)),
			Language: "en",
		}
		err := s.Repo.Save(ctx, testWord)
		assert.NoError(t, err)
	}

	// Execute with limit 3
	filter := map[string]interface{}{"language": "en"}
	words, err := s.Repo.List(ctx, filter, 3, 0)

	// Assert
	assert.NoError(t, err)
	assert.LessOrEqual(t, len(words), 3)
}
