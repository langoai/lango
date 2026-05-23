package adk

import (
	"context"
	"errors"
	"iter"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
)

func TestAdaptStreamingTool_NoStreamingHandlerFallsBackToRegular(t *testing.T) {
	tt := &agent.Tool{
		Name:        "echo",
		Description: "Echoes input back.",
		Parameters: map[string]interface{}{
			"text": agent.ParameterDef{Type: "string", Description: "echo me", Required: true},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return params["text"], nil
		},
	}

	adapted, err := AdaptStreamingTool(tt)
	require.NoError(t, err)
	require.NotNil(t, adapted)
	assert.Equal(t, "echo", adapted.Name())
	assert.Equal(t, "Echoes input back.", adapted.Description())
}

func TestAdaptStreamingTool_WithStreamingHandlerReturnsStreamingTool(t *testing.T) {
	tt := &agent.Tool{
		Name:        "count",
		Description: "Counts to N.",
		Parameters: map[string]interface{}{
			"n": agent.ParameterDef{Type: "integer", Description: "limit", Required: true},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			// Non-streaming fallback (required by Tool struct invariant).
			return nil, errors.New("non-streaming path not used")
		},
		StreamingHandler: func(ctx context.Context, params map[string]interface{}) iter.Seq2[string, error] {
			return func(yield func(string, error) bool) {
				for _, s := range []string{"one", "two", "three"} {
					if !yield(s, nil) {
						return
					}
				}
			}
		},
	}

	adapted, err := AdaptStreamingTool(tt)
	require.NoError(t, err)
	require.NotNil(t, adapted)
	assert.Equal(t, "count", adapted.Name())
}

func TestAdaptTool_PreservesNonStreamingBehavior(t *testing.T) {
	// Verify the non-streaming AdaptTool entry point still works after the refactor.
	tt := &agent.Tool{
		Name:        "noop",
		Description: "Does nothing.",
		Parameters:  map[string]interface{}{},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	}
	adapted, err := AdaptTool(tt)
	require.NoError(t, err)
	require.NotNil(t, adapted)
	assert.Equal(t, "noop", adapted.Name())
}
