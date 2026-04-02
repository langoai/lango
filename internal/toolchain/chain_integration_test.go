package toolchain

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
)

// TestMiddlewareChain_OutputManager_LargeResult verifies that the output
// manager middleware triggers compression for large tool results and that
// the compressed output contains expected metadata and a stored reference
// when a store is provided.
func TestMiddlewareChain_OutputManager_LargeResult(t *testing.T) {
	t.Parallel()

	// Generate text that exceeds 3x the token budget (large tier).
	// With budget=100 tokens and ~4 chars/token, we need >1200 chars for >300 tokens.
	makeText := func(tokens int) string {
		lineTokens := 10
		numLines := tokens / lineTokens
		if numLines < 1 {
			numLines = 1
		}
		var sb strings.Builder
		for i := 0; i < numLines; i++ {
			sb.WriteString(strings.Repeat("abcd", lineTokens))
			if i < numLines-1 {
				sb.WriteByte('\n')
			}
		}
		return sb.String()
	}

	tests := []struct {
		give          string
		budget        int
		resultTokens  int
		withStore     bool
		wantTier      string
		wantCompress  bool
		wantStoredRef bool
	}{
		{
			give:         "small result stays uncompressed",
			budget:       2000,
			resultTokens: 50,
			wantTier:     "small",
			wantCompress: false,
		},
		{
			give:         "medium result gets compressed",
			budget:       100,
			resultTokens: 200,
			wantTier:     "medium",
			wantCompress: true,
		},
		{
			give:         "large result gets aggressively compressed without store",
			budget:       100,
			resultTokens: 500,
			wantTier:     "large",
			wantCompress: true,
		},
		{
			give:          "large result gets stored when store available",
			budget:        100,
			resultTokens:  500,
			withStore:     true,
			wantTier:      "large",
			wantCompress:  true,
			wantStoredRef: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			largeText := makeText(tt.resultTokens)
			tool := &agent.Tool{
				Name: "test_tool",
				Handler: func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
					return largeText, nil
				},
			}

			cfg := config.OutputManagerConfig{
				TokenBudget: tt.budget,
				HeadRatio:   0.7,
				TailRatio:   0.3,
			}

			var store *fakeStore
			var mw Middleware
			if tt.withStore {
				store = &fakeStore{ref: "ref-" + tt.give}
				mw = WithOutputManager(cfg, store)
			} else {
				mw = WithOutputManager(cfg)
			}

			wrapped := Chain(tool, mw)
			result, err := wrapped.Handler(context.Background(), nil)
			require.NoError(t, err)

			m, ok := result.(map[string]interface{})
			require.True(t, ok, "result should be a map with _meta")

			meta, hasMeta := m["_meta"]
			require.True(t, hasMeta, "result should have _meta field")
			metaMap, ok := meta.(map[string]interface{})
			require.True(t, ok, "_meta should be a map")

			assert.Equal(t, tt.wantTier, metaMap["tier"])
			assert.Equal(t, tt.wantCompress, metaMap["compressed"])

			if tt.wantCompress {
				content, hasContent := m["content"]
				require.True(t, hasContent, "compressed result should have content field")
				contentStr, ok := content.(string)
				require.True(t, ok)
				assert.Contains(t, contentStr, "[compressed: removed",
					"compressed content should contain compression marker")
			}

			if tt.wantStoredRef {
				require.NotNil(t, store, "store should be provided")
				assert.Equal(t, "ref-"+tt.give, metaMap["storedRef"])
				assert.Equal(t, "test_tool", store.lastToolName)
				assert.NotEmpty(t, store.lastContent, "store should have received the full content")
			}
		})
	}
}

// TestMiddlewareChain_LearningObserver_Records verifies that the learning
// observer middleware records tool execution events including tool name,
// parameters, result, and errors.
func TestMiddlewareChain_LearningObserver_Records(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give       string
		toolName   string
		params     map[string]interface{}
		result     interface{}
		toolErr    error
		wantCalls  int
		wantResult interface{}
	}{
		{
			give:       "successful execution is recorded",
			toolName:   "fs_read",
			params:     map[string]interface{}{"path": "/tmp/test.txt"},
			result:     "file contents here",
			wantCalls:  1,
			wantResult: "file contents here",
		},
		{
			give:       "error execution is recorded",
			toolName:   "exec",
			params:     map[string]interface{}{"command": "bad-cmd"},
			result:     nil,
			toolErr:    assert.AnError,
			wantCalls:  1,
			wantResult: nil,
		},
		{
			give:       "nil params are handled",
			toolName:   "my_tool",
			params:     nil,
			result:     "ok",
			wantCalls:  1,
			wantResult: "ok",
		},
		{
			give:       "complex result is recorded",
			toolName:   "api_call",
			params:     map[string]interface{}{"url": "https://example.com", "method": "GET"},
			result:     map[string]interface{}{"status": 200, "body": "ok"},
			wantCalls:  1,
			wantResult: map[string]interface{}{"status": 200, "body": "ok"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			obs := &mockObserver{}
			mw := WithLearning(obs)

			tool := &agent.Tool{
				Name: tt.toolName,
				Handler: func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
					return tt.result, tt.toolErr
				},
			}

			wrapped := Chain(tool, mw)
			result, err := wrapped.Handler(context.Background(), tt.params)

			if tt.toolErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.toolErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}

			require.Len(t, obs.calls, tt.wantCalls)
			assert.Equal(t, tt.toolName, obs.calls[0].toolName)
			assert.Equal(t, tt.wantResult, obs.calls[0].result)
			if tt.toolErr != nil {
				assert.Equal(t, tt.toolErr, obs.calls[0].err)
			} else {
				assert.NoError(t, obs.calls[0].err)
			}
		})
	}
}

// TestMiddlewareChain_LearningObserver_MultipleInvocations verifies the
// observer accumulates records across multiple tool invocations.
func TestMiddlewareChain_LearningObserver_MultipleInvocations(t *testing.T) {
	t.Parallel()

	obs := &mockObserver{}
	mw := WithLearning(obs)

	callCount := 0
	tool := &agent.Tool{
		Name: "counter_tool",
		Handler: func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
			callCount++
			return callCount, nil
		},
	}

	wrapped := Chain(tool, mw)

	for i := 0; i < 5; i++ {
		_, err := wrapped.Handler(context.Background(), nil)
		require.NoError(t, err)
	}

	require.Len(t, obs.calls, 5, "observer should record each invocation")
	for i, call := range obs.calls {
		assert.Equal(t, "counter_tool", call.toolName)
		assert.Equal(t, i+1, call.result, "each call should record its unique result")
	}
}

// TestMiddlewareChain_OutputManagerWithLearning_Integration verifies that
// output manager and learning observer work together in the correct order.
// Learning observes the raw result (before output management), then output
// manager compresses it. This is the production order because learning is
// applied first (innermost), output manager second.
func TestMiddlewareChain_OutputManagerWithLearning_Integration(t *testing.T) {
	t.Parallel()

	// Generate large text.
	largeText := strings.Repeat("abcd", 300) // ~300 tokens

	obs := &mockObserver{}
	learningMW := WithLearning(obs)

	cfg := config.OutputManagerConfig{TokenBudget: 100, HeadRatio: 0.7, TailRatio: 0.3}
	outputMW := WithOutputManager(cfg)

	tool := &agent.Tool{
		Name: "big_tool",
		Handler: func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
			return largeText, nil
		},
	}

	// Apply in production order: learning first (innermost), output manager second.
	tools := []*agent.Tool{tool}
	tools = ChainAll(tools, learningMW)
	tools = ChainAll(tools, outputMW)

	result, err := tools[0].Handler(context.Background(), nil)
	require.NoError(t, err)

	// Output manager should have compressed the result.
	m, ok := result.(map[string]interface{})
	require.True(t, ok)
	meta, ok := m["_meta"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, meta["compressed"], "output manager should compress large result")

	// Learning observer sees the raw result (before output manager processes it),
	// because learning is innermost and executes after the handler but before
	// output manager's post-processing.
	require.Len(t, obs.calls, 1)
	assert.Equal(t, largeText, obs.calls[0].result,
		"learning observer should see the raw uncompressed result")
}
