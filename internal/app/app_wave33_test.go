package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/testutil"
)

func TestWave33InitTurnTraceStoreRequiresEntSessionStore(t *testing.T) {
	t.Parallel()

	assert.Nil(t, initTurnTraceStore(nil))
	assert.Nil(t, initTurnTraceStore(&stubSessionStore{}))

	entStore := session.NewEntStoreWithClient(testutil.TestEntClient(t))
	got := initTurnTraceStore(entStore)

	require.NotNil(t, got)
}
