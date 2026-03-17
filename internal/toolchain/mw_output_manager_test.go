package toolchain

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/tooloutput"
)

func boolPtr(b bool) *bool { return &b }

// fakeStore implements OutputStorer for testing.
type fakeStore struct {
	lastToolName string
	lastContent  string
	ref          string
}

func (s *fakeStore) Store(toolName, content string) string {
	s.lastToolName = toolName
	s.lastContent = content
	return s.ref
}

func TestWithOutputManager(t *testing.T) {
	t.Parallel()

	// Generate a multi-line string with approximately the given number of tokens.
	// ASCII: ~4 chars per token. Each line is ~10 tokens (40 chars).
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
		give       string
		cfg        config.OutputManagerConfig
		result     interface{}
		wantErr    error
		wantTier   string
		wantCompr  bool
		wantPassth bool // true if result should pass through unchanged
	}{
		{
			give:      "small output under budget",
			cfg:       config.OutputManagerConfig{TokenBudget: 2000},
			result:    "short text",
			wantTier:  tierSmall,
			wantCompr: false,
		},
		{
			give:      "medium output compressed",
			cfg:       config.OutputManagerConfig{TokenBudget: 100, HeadRatio: 0.7, TailRatio: 0.3},
			result:    makeText(200), // 200 tokens, budget 100 → medium tier
			wantTier:  tierMedium,
			wantCompr: true,
		},
		{
			give:      "large output aggressively compressed",
			cfg:       config.OutputManagerConfig{TokenBudget: 100, HeadRatio: 0.7, TailRatio: 0.3},
			result:    makeText(500), // 500 tokens, budget 100 → large tier (>3x)
			wantTier:  tierLarge,
			wantCompr: true,
		},
		{
			give:       "disabled config passes through unchanged",
			cfg:        config.OutputManagerConfig{Enabled: boolPtr(false), TokenBudget: 10},
			result:     makeText(500),
			wantPassth: true,
		},
		{
			give:      "map result gets meta injected",
			cfg:       config.OutputManagerConfig{TokenBudget: 2000},
			result:    map[string]interface{}{"key": "value"},
			wantTier:  tierSmall,
			wantCompr: false,
		},
		{
			give:       "error result passes through unchanged",
			cfg:        config.OutputManagerConfig{TokenBudget: 100},
			result:     "some result",
			wantErr:    errors.New("tool error"),
			wantPassth: true,
		},
		{
			give:       "nil result passes through unchanged",
			cfg:        config.OutputManagerConfig{TokenBudget: 100},
			result:     nil,
			wantPassth: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			tool := &agent.Tool{Name: "test_tool"}
			handler := func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				return tt.result, tt.wantErr
			}

			mw := WithOutputManager(tt.cfg)
			wrapped := mw(tool, handler)
			got, err := wrapped(context.Background(), nil)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.result, got, "error result should pass through unchanged")
				return
			}

			require.NoError(t, err)

			if tt.wantPassth {
				assert.Equal(t, tt.result, got, "pass-through result should be unchanged")
				return
			}

			// Verify _meta is present.
			m, ok := got.(map[string]interface{})
			require.True(t, ok, "result should be a map with _meta")

			meta, hasMeta := m["_meta"]
			require.True(t, hasMeta, "result should have _meta field")

			metaMap, ok := meta.(map[string]interface{})
			require.True(t, ok, "_meta should be a map")

			assert.Equal(t, tt.wantTier, metaMap["tier"])
			assert.Equal(t, tt.wantCompr, metaMap["compressed"])
			assert.NotNil(t, metaMap["originalTokens"])
			assert.NotEmpty(t, metaMap["contentType"])

			if tt.wantTier == tierLarge {
				_, hasRef := metaMap["storedRef"]
				assert.True(t, hasRef, "large tier should have storedRef")
				// Without a store, storedRef is nil.
				assert.Nil(t, metaMap["storedRef"], "storedRef should be nil without store")
			}

			if tt.wantCompr {
				// For string results, check content field contains compression marker.
				if content, hasContent := m["content"]; hasContent {
					s, isStr := content.(string)
					if isStr {
						assert.Contains(t, s, "[compressed: removed")
					}
				}
			}
		})
	}
}

func TestWithOutputManager_WithStore(t *testing.T) {
	t.Parallel()

	// Generate large text that exceeds 3x budget.
	makeText := func(tokens int) string {
		lineTokens := 10
		numLines := tokens / lineTokens
		var sb strings.Builder
		for i := 0; i < numLines; i++ {
			sb.WriteString(strings.Repeat("abcd", lineTokens))
			if i < numLines-1 {
				sb.WriteByte('\n')
			}
		}
		return sb.String()
	}

	store := &fakeStore{ref: "test-ref-123"}
	cfg := config.OutputManagerConfig{TokenBudget: 100, HeadRatio: 0.7, TailRatio: 0.3}

	tool := &agent.Tool{Name: "big_tool"}
	handler := func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		return makeText(500), nil
	}

	mw := WithOutputManager(cfg, store)
	wrapped := mw(tool, handler)
	got, err := wrapped(context.Background(), nil)

	require.NoError(t, err)

	m, ok := got.(map[string]interface{})
	require.True(t, ok)

	metaMap, ok := m["_meta"].(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, tierLarge, metaMap["tier"])
	assert.Equal(t, "test-ref-123", metaMap["storedRef"])
	assert.Equal(t, "big_tool", store.lastToolName)
	assert.NotEmpty(t, store.lastContent)
}

func TestWithOutputManager_DefaultConfig(t *testing.T) {
	t.Parallel()

	// Zero-value config should use defaults and be enabled.
	cfg := config.OutputManagerConfig{}
	mw := WithOutputManager(cfg)

	tool := &agent.Tool{Name: "test_tool"}
	handler := func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		return "hello", nil
	}
	wrapped := mw(tool, handler)
	got, err := wrapped(context.Background(), nil)

	require.NoError(t, err)
	m, ok := got.(map[string]interface{})
	require.True(t, ok, "result should be a map with _meta")
	_, hasMeta := m["_meta"]
	assert.True(t, hasMeta, "default config should enable output management")
}

func TestInjectMeta_StringResult(t *testing.T) {
	t.Parallel()

	meta := outputMeta{
		OriginalTokens: 100,
		Tier:           tierSmall,
		ContentType:    tooloutput.ContentTypeText,
		Compressed:     false,
	}

	got := injectMeta("hello world", true, meta)
	m, ok := got.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "hello world", m["content"])

	metaMap, ok := m["_meta"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 100, metaMap["originalTokens"])
	assert.Equal(t, tierSmall, metaMap["tier"])
}

func TestInjectMeta_MapResult(t *testing.T) {
	t.Parallel()

	meta := outputMeta{
		OriginalTokens: 50,
		Tier:           tierSmall,
		ContentType:    tooloutput.ContentTypeJSON,
		Compressed:     false,
	}

	input := map[string]interface{}{"key": "value"}
	got := injectMeta(input, false, meta)
	m, ok := got.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "value", m["key"])

	_, hasMeta := m["_meta"]
	assert.True(t, hasMeta, "_meta should be injected directly into map")
}
