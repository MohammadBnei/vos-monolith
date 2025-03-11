package server

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"voconsteroid/internal/domain/word"
)

// WordSearchRequest represents a request to search for a word
type WordSearchRequest struct {
	Text     string `json:"text" binding:"required"`
	Language string `json:"language" binding:"required"`
}

// WordSearchResponse represents the response for a word search
type WordSearchResponse struct {
	Word *word.Word `json:"word"`
}

// RecentWordsRequest represents a request for recent words
type RecentWordsRequest struct {
	Language string `form:"language"`
	Limit    int    `form:"limit"`
}

// RecentWordsResponse represents the response for recent words
type RecentWordsResponse struct {
	Words []*word.Word `json:"words"`
}

// RelatedWordsRequest represents a request for related words
type RelatedWordsRequest struct {
	WordID string `uri:"wordId" binding:"required"`
}

// RelatedWordsResponse represents the response for related words
type RelatedWordsResponse struct {
	SourceWord *word.Word   `json:"source_word"`
	Synonyms   []*word.Word `json:"synonyms,omitempty"`
	Antonyms   []*word.Word `json:"antonyms,omitempty"`
}

// AutoCompleteRequest represents a request for autocomplete suggestions
type AutoCompleteRequest struct {
	Prefix   string `form:"q" binding:"required,min=2"`
	Language string `form:"lang" binding:"omitempty,oneof=fr en"`
}

// AutoCompleteResponse represents the response for autocomplete suggestions
type AutoCompleteResponse struct {
	Suggestions []string `json:"suggestions"`
}

// SearchWord handles word search requests
// @Summary Search for a word
// @Description Search for a word by text and language
// @Tags words
// @Accept json
// @Produce json
// @Param request body WordSearchRequest true "Word search request"
// @Success 200 {object} WordSearchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/words/search [post]
func (s *Server) SearchWord(c *gin.Context) {
	log := c.MustGet("logger").(zerolog.Logger)

	var req WordSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Debug().Err(err).Msg("Invalid word search request")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	log.Debug().Str("text", req.Text).Str("language", req.Language).Msg("Searching for word")

	foundWord, err := s.wordService.Search(c.Request.Context(), req.Text, req.Language)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to search for word"

		if errors.Is(err, word.ErrWordNotFound) {
			status = http.StatusNotFound
			message = "Word not found"
		} else if errors.Is(err, word.ErrInvalidWord) {
			status = http.StatusBadRequest
			message = "Invalid word"
		}

		log.Debug().Err(err).Str("word", req.Text).Str("language", req.Language).Msg(message)
		c.JSON(status, ErrorResponse{
			Status:  status,
			Message: message,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, WordSearchResponse{
		Word: foundWord,
	})
}

// GetRecentWords handles requests for recent words
// @Summary Get recent words
// @Description Get recently searched words
// @Tags words
// @Accept json
// @Produce json
// @Param language query string false "Language code" Enums(fr, en) default(fr)
// @Param limit query integer false "Maximum number of words to return" minimum(1) maximum(100) default(10)
// @Success 200 {object} RecentWordsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/words/recent [get]
func (s *Server) GetRecentWords(c *gin.Context) {
	log := c.MustGet("logger").(zerolog.Logger)

	// Get parameters from query
	language := c.Query("language")
	if language == "" {
		language = "fr" // Default to French
	}
	
	limitStr := c.Query("limit")
	limit := 10 // Default limit
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			log.Debug().Err(err).Str("limit", limitStr).Msg("Invalid limit parameter")
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Status:  http.StatusBadRequest,
				Message: "Invalid limit parameter",
				Error:   "Limit must be between 1 and 100",
			})
			return
		}
	}

	log.Debug().Str("language", language).Int("limit", limit).Msg("Getting recent words")

	words, err := s.wordService.GetRecentWords(c.Request.Context(), language, limit)
	if err != nil {
		log.Error().Err(err).Msg("Error getting recent words")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Error getting recent words",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RecentWordsResponse{
		Words: words,
	})
}

// AutoComplete handles autocomplete requests
// @Summary Get autocomplete suggestions
// @Description Get word suggestions based on prefix
// @Tags words
// @Accept json
// @Produce json
// @Param q query string true "The prefix to search for" minlength(2)
// @Param lang query string false "Language code" Enums(fr, en)
// @Success 200 {object} AutoCompleteResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/words/autocomplete [get]
func (s *Server) AutoComplete(c *gin.Context) {
	log := c.MustGet("logger").(zerolog.Logger)

	var req AutoCompleteRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Debug().Err(err).Msg("Invalid autocomplete request")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	// Default to French if no language specified
	language := req.Language
	if language == "" {
		language = "fr"
	}

	log.Debug().Str("prefix", req.Prefix).Str("language", language).Msg("Getting autocomplete suggestions")

	suggestions, err := s.wordService.GetSuggestions(c.Request.Context(), req.Prefix, language)
	if err != nil {
		log.Debug().Err(err).Str("prefix", req.Prefix).Str("language", req.Language).Msg("Failed to get autocomplete suggestions")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to get suggestions",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, AutoCompleteResponse{
		Suggestions: suggestions,
	})
}

// GetRelatedWords handles requests for related words
// @Summary Get related words
// @Description Get synonyms and antonyms for a word
// @Tags words
// @Accept json
// @Produce json
// @Param wordId path string true "Word ID"
// @Success 200 {object} RelatedWordsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/words/{wordId}/related [get]
func (s *Server) GetRelatedWords(c *gin.Context) {
	log := c.MustGet("logger").(zerolog.Logger)

	wordId := c.Param("wordId")
	if wordId == "" {
		log.Debug().Msg("Missing wordId parameter")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Missing wordId parameter",
			Error:   "wordId is required",
		})
		return
	}

	log.Debug().Str("wordId", wordId).Msg("Getting related words")

	relatedWords, err := s.wordService.GetRelatedWords(c.Request.Context(), wordId)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, word.ErrWordNotFound) {
			status = http.StatusNotFound
		}

		log.Debug().Err(err).Msg("Error getting related words")
		c.JSON(status, ErrorResponse{
			Status:  status,
			Message: "Error getting related words",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RelatedWordsResponse{
		SourceWord: relatedWords.SourceWord,
		Synonyms:   relatedWords.Synonyms,
		Antonyms:   relatedWords.Antonyms,
	})
}
