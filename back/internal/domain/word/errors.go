package word

import "errors"

// Domain errors
var (
	ErrInvalidWordType = errors.New("invalid word type")
	ErrInvalidGender   = errors.New("invalid gender")
	ErrWordNotFound    = errors.New("word not found")
	ErrInvalidWord     = errors.New("invalid word")
	ErrDuplicateWord   = errors.New("word already exists")
)
