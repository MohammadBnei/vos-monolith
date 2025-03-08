package server

import (
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

// RecentWordsResponse represents the response for recent words
type RecentWordsResponse struct {
	Words []*word.Word `json:"words"`
}

// searchWord handles requests to search for a word
func (s *Server) searchWord(c *gin.Context) {
	log := c.MustGet("logger").(zerolog.Logger)
	
	var req WordSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Debug().Err(err).Msg("Invalid search word request")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}
	
	// Validate language
	if req.Language == "" {
		req.Language = "en" // Default to English
	}
	
	// Search for the word
	foundWord, err := s.wordService.Search(c.Request.Context(), req.Text, req.Language)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to search for word"
		
		if err == word.ErrWordNotFound {
			status = http.StatusNotFound
			message = "Word not found"
		} else if err == word.ErrInvalidWord {
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
	
	// Return the word
	c.JSON(http.StatusOK, WordSearchResponse{
		Word: foundWord,
	})
}

// getRecentWords handles requests to get recently searched words
func (s *Server) getRecentWords(c *gin.Context) {
	log := c.MustGet("logger").(zerolog.Logger)
	
	language := c.Query("language")
	if language == "" {
		language = "en" // Default to English
	}
	
	// Get recent words
	recentWords, err := s.wordService.GetRecentWords(c.Request.Context(), language, 10)
	if err != nil {
		log.Debug().Err(err).Str("language", language).Msg("Failed to get recent words")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to get recent words",
			Error:   err.Error(),
		})
		return
	}
	
	// Return the words
	c.JSON(http.StatusOK, RecentWordsResponse{
		Words: recentWords,
	})
}
