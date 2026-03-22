package provenance

import "errors"

// Sentinel errors for the provenance package.
var (
	ErrCheckpointNotFound = errors.New("checkpoint not found")
	ErrSessionNotFound    = errors.New("session node not found")
	ErrMaxCheckpoints     = errors.New("maximum checkpoints per session reached")
	ErrInvalidLabel       = errors.New("checkpoint label is required")
	ErrInvalidRunID       = errors.New("run ID is required")
	ErrInvalidSessionKey  = errors.New("session key is required")
	ErrInvalidRedaction   = errors.New("invalid redaction level")
)
