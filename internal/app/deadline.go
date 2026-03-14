package app

import (
	"context"
	"time"

	"github.com/langoai/lango/internal/deadline"
)

// ExtendableDeadline is an alias for backward compatibility within the app package.
type ExtendableDeadline = deadline.ExtendableDeadline

// NewExtendableDeadline creates a new ExtendableDeadline.
// Deprecated: Use deadline.New directly.
func NewExtendableDeadline(parent context.Context, baseTimeout, maxTimeout time.Duration) (context.Context, *ExtendableDeadline) {
	return deadline.New(parent, baseTimeout, maxTimeout)
}
