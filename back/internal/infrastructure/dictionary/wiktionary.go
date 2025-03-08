package dictionary

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"voconsteroid/internal/domain/word"
)

// WiktionaryAPI implements the word.DictionaryAPI interface for Wiktionary
type WiktionaryAPI struct {
	client  *http.Client
	baseURL string
	logger  zerolog.Logger
}

// NewWiktionaryAPI creates a new Wiktionary API client
func NewWiktionaryAPI(logger zerolog.Logger) *WiktionaryAPI {
	return &WiktionaryAPI{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://en.wiktionary.org/w/api.php",
		logger:  logger.With().Str("component", "wiktionary_api").Logger(),
	}
}

// FetchWord retrieves word information from Wiktionary
func (w *WiktionaryAPI) FetchWord(ctx context.Context, text, language string) (*word.Word, error) {
	w.logger.Debug().Str("text", text).Str("language", language).Msg("Fetching word from Wiktionary")
	
	// Build URL for the API request
	url := fmt.Sprintf("%s?action=query&format=json&prop=extracts|translations|pronunciation&titles=%s",
		w.baseURL, text)

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		w.logger.Error().Err(err).Str("url", url).Msg("Failed to create request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := w.client.Do(req)
	if err != nil {
		w.logger.Error().Err(err).Str("url", url).Msg("Failed to send request")
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		w.logger.Error().Int("status", resp.StatusCode).Str("url", url).Msg("Received non-200 response")
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	// Parse response
	var response wiktionaryResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		w.logger.Error().Err(err).Msg("Failed to parse response")
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Process response
	newWord := word.NewWord(text, language)
	
	// Ensure maps and slices are initialized
	if newWord.Definitions == nil {
		newWord.Definitions = []string{}
	}
	if newWord.Examples == nil {
		newWord.Examples = []string{}
	}
	if newWord.Translations == nil {
		newWord.Translations = make(map[string]string)
	}

	// Extract data from response
	for _, page := range response.Query.Pages {
		if page.Extract != "" {
			// Simple parsing of the extract to get definitions
			// In a real implementation, you would want more sophisticated parsing
			lines := strings.Split(page.Extract, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "<") {
					newWord.Definitions = append(newWord.Definitions, line)
				}
			}
		}

		// Add pronunciation if available
		if page.Pronunciation != "" {
			newWord.Pronunciation = page.Pronunciation
		}

		// Add translations if available
		if page.Translations != nil {
			for lang, translation := range page.Translations {
				newWord.Translations[lang] = translation
			}
		}
	}

	// If no definitions were found, return an error
	if len(newWord.Definitions) == 0 {
		w.logger.Warn().Str("text", text).Str("language", language).Msg("No word data found")
		return nil, fmt.Errorf("no word data found: %w", word.ErrWordNotFound)
	}

	w.logger.Debug().
		Str("text", text).
		Str("language", language).
		Int("definitions", len(newWord.Definitions)).
		Int("translations", len(newWord.Translations)).
		Msg("Successfully fetched word data")
		
	return newWord, nil
}

// wiktionaryResponse represents the response structure from Wiktionary API
type wiktionaryResponse struct {
	Query struct {
		Pages map[string]struct {
			Extract       string            `json:"extract"`
			Pronunciation string            `json:"pronunciation"`
			Translations  map[string]string `json:"translations"`
		} `json:"pages"`
	} `json:"query"`
}
