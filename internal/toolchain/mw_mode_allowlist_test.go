package toolchain

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func okHandler(_ context.Context, _ map[string]interface{}) (interface{}, error) {
	return "ok", nil
}

func TestWithModeAllowlist_AllowedToolPasses(t *testing.T) {
	tool := makeTool("allowed_tool", okHandler)
	resolver := func(ctx context.Context) (map[string]bool, bool) {
		return map[string]bool{"allowed_tool": true}, true
	}
	wrapped := Chain(tool, WithModeAllowlist(resolver))
	res, err := wrapped.Handler(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, "ok", res)
}

func TestWithModeAllowlist_BlockedToolReturnsError(t *testing.T) {
	tool := makeTool("blocked_tool", okHandler)
	resolver := func(ctx context.Context) (map[string]bool, bool) {
		return map[string]bool{"other_tool": true}, true
	}
	wrapped := Chain(tool, WithModeAllowlist(resolver))
	_, err := wrapped.Handler(context.Background(), nil)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "blocked_tool") &&
		strings.Contains(err.Error(), "not available in current mode"),
		"error should reference blocked tool and mode constraint, got: %v", err)
}

func TestWithModeAllowlist_NoActiveModePassesThrough(t *testing.T) {
	tool := makeTool("any_tool", okHandler)
	resolver := func(ctx context.Context) (map[string]bool, bool) {
		return nil, false
	}
	wrapped := Chain(tool, WithModeAllowlist(resolver))
	res, err := wrapped.Handler(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, "ok", res)
}

func TestWithModeAllowlist_NilResolverPassesThrough(t *testing.T) {
	tool := makeTool("any_tool", okHandler)
	wrapped := Chain(tool, WithModeAllowlist(nil))
	res, err := wrapped.Handler(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, "ok", res)
}
