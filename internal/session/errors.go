package session

import "errors"

var (
	ErrSessionNotFound  = errors.New("session not found")
	ErrDuplicateSession = errors.New("duplicate session")
)
