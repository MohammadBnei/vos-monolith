package server

import (
	"errors"
	"net/http"

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
	Language string `json:"language" binding:"required"`
	Limit    int    `json:"limit,omitempty"`
}

// RecentWordsResponse represents the response for recent words
type RecentWordsResponse struct {
	Words []*word.Word `json:"words"`
}

// RelatedWordsRequest represents a request for related words
type RelatedWordsRequest struct {
	WordID string `uri:"id" binding:"required"`
}

// RelatedWordsResponse represents the response for related words
type RelatedWordsResponse struct {
	SourceWord *word.Word   `json:"source_word"`
	Synonyms   []*word.Word `json:"synonyms,omitempty"`
	Antonyms   []*word.Word `json:"antonyms,omitempty"`
}

// AutoCompleteRequest represents a request for autocomplete suggestions
type AutoCompleteRequest struct {
	Prefix   string `form:"prefix" binding:"required"`
	Language string `form:"language" binding:"required"`
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
// @Param request body RecentWordsRequest true "Recent words request"
// @Success 200 {object} RecentWordsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/words/recent [post]
func (s *Server) GetRecentWords(c *gin.Context) {
	log := c.MustGet("logger").(zerolog.Logger)

	var req RecentWordsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Debug().Err(err).Msg("Invalid recent words request")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	// Default limit if not provided
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	log.Debug().Str("language", req.Language).Int("limit", limit).Msg("Getting recent words")

	words, err := s.wordService.GetRecentWords(c.Request.Context(), req.Language, limit)
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
// @Param prefix query string true "Word prefix"
// @Param language query string true "Language code"
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

	log.Debug().Str("prefix", req.Prefix).Str("language", req.Language).Msg("Getting autocomplete suggestions")

	suggestions, err := s.wordService.GetSuggestions(c.Request.Context(), req.Prefix, req.Language)
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
// @Param id path string true "Word ID"
// @Success 200 {object} RelatedWordsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/words/{id}/related [get]
func (s *Server) GetRelatedWords(c *gin.Context) {
	log := c.MustGet("logger").(zerolog.Logger)

	var req RelatedWordsRequest
	if err := c.ShouldBindUri(&req); err != nil {
		log.Debug().Err(err).Msg("Invalid related words request")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	log.Debug().Str("wordID", req.WordID).Msg("Getting related words")

	relatedWords, err := s.wordService.GetRelatedWords(c.Request.Context(), req.WordID)
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
