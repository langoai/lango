package memory

import (
	"time"

	"github.com/google/uuid"
)

// Observation represents a compressed note from conversation history.
type Observation struct {
	ID               uuid.UUID
	SessionKey       string
	Content          string
	TokenCount       int
	SourceStartIndex int
	SourceEndIndex   int
	CreatedAt        time.Time
}

// Reflection represents condensed observations.
type Reflection struct {
	ID         uuid.UUID
	SessionKey string
	Content    string
	TokenCount int
	Generation int
	CreatedAt  time.Time
}
