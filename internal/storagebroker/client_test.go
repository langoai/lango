package storagebroker

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewBrokerCommandUsesInjectedStderr(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	cmd, stdin, stdout, err := newBrokerCommand(context.Background(), "/bin/echo", &stderr)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, stdin.Close())
		require.NoError(t, stdout.Close())
	})

	require.Same(t, &stderr, cmd.Stderr)
}
