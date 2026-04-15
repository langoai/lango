package session

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModeNameFromContext_EmptyWhenUnset(t *testing.T) {
	got := ModeNameFromContext(context.Background())
	assert.Equal(t, "", got)
}

func TestWithModeName_RoundTrip(t *testing.T) {
	ctx := WithModeName(context.Background(), "code-review")
	assert.Equal(t, "code-review", ModeNameFromContext(ctx))
}

func TestWithModeName_OverwritesExisting(t *testing.T) {
	ctx := WithModeName(context.Background(), "research")
	ctx = WithModeName(ctx, "debug")
	assert.Equal(t, "debug", ModeNameFromContext(ctx))
}

func TestSession_ModeMetadata(t *testing.T) {
	s := &Session{}
	assert.Equal(t, "", s.Mode(), "initial mode should be empty")

	s.SetMode("research")
	assert.Equal(t, "research", s.Mode())
	assert.Equal(t, "research", s.Metadata[MetadataKeyMode])

	s.SetMode("")
	assert.Equal(t, "", s.Mode(), "empty SetMode clears the mode")
	_, stillPresent := s.Metadata[MetadataKeyMode]
	assert.False(t, stillPresent, "mode key should be removed when cleared")
}
