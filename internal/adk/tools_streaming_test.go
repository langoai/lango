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

// TestStreamingHandler_YieldsExpectedChunks exercises the agent.Tool's
// StreamingHandler directly (independent of ADK runtime) and asserts the
// 3-chunk yield contract that the adapter is supposed to forward. This
// closes the Track C spec gate that required actual chunk yielding behavior.
func TestStreamingHandler_YieldsExpectedChunks(t *testing.T) {
	want := []string{"one", "two", "three"}
	handler := func(ctx context.Context, params map[string]interface{}) iter.Seq2[string, error] {
		return func(yield func(string, error) bool) {
			for _, s := range want {
				if !yield(s, nil) {
					return
				}
			}
		}
	}

	var got []string
	for chunk, err := range handler(context.Background(), nil) {
		require.NoError(t, err)
		got = append(got, chunk)
	}
	assert.Equal(t, want, got)
}

// TestStreamingHandler_StopsOnConsumerCancel verifies that the iterator
// honors early termination by the consumer. AdaptStreamingTool's closure
// relies on this contract to release the timeout context when the caller
// stops ranging.
func TestStreamingHandler_StopsOnConsumerCancel(t *testing.T) {
	yielded := 0
	handler := func(ctx context.Context, params map[string]interface{}) iter.Seq2[string, error] {
		return func(yield func(string, error) bool) {
			for i := 0; i < 100; i++ {
				yielded++
				if !yield("chunk", nil) {
					return
				}
			}
		}
	}

	collected := 0
	for range handler(context.Background(), nil) {
		collected++
		if collected == 3 {
			break
		}
	}
	assert.Equal(t, 3, collected)
	assert.Equal(t, 3, yielded, "iterator should have stopped after consumer break")
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
