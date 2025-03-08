package word

import "errors"

// Domain errors
var (
	ErrWordNotFound  = errors.New("word not found")
	ErrInvalidWord   = errors.New("invalid word")
	ErrDuplicateWord = errors.New("word already exists")
)
