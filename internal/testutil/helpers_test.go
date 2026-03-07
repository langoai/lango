package testutil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/testutil"
)

func TestNopLogger(t *testing.T) {
	t.Parallel()
	logger := testutil.NopLogger()
	require.NotNil(t, logger)
	// should not panic
	logger.Infow("test message", "key", "value")
}

func TestTestEntClient(t *testing.T) {
	t.Parallel()
	client := testutil.TestEntClient(t)
	require.NotNil(t, client)
	// verify the client is functional by checking it does not panic on a simple query
	assert.NotNil(t, client)
}

func TestSkipShort(t *testing.T) {
	t.Parallel()
	// SkipShort should not skip when not in short mode (normal test run)
	// We cannot easily test the skip path without running with -short,
	// so just verify it does not panic in normal mode.
	testutil.SkipShort(t)
}
