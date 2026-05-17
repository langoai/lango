package bg

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/background"
)

type stubRunner struct {
	result string
	delay  time.Duration
}

func (s stubRunner) Run(ctx context.Context, sessionKey string, prompt string) (string, error) {
	if s.delay > 0 {
		select {
		case <-time.After(s.delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return s.result, nil
}

func executeBgCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestBgHelpExplainsInProcessManagerScope(t *testing.T) {
	cmd := NewBgCmd(func() (*background.Manager, error) {
		t.Fatal("help should not resolve the manager")
		return nil, nil
	})

	out, err := executeBgCommand(t, cmd, "--help")

	require.NoError(t, err)
	assert.Contains(t, out, "supplied in-process manager")
	assert.Contains(t, out, "Embedded callers")
}

func TestBgList_WritesToCommandOutput(t *testing.T) {
	mgr := background.NewManager(stubRunner{result: "done"}, nil, 5, time.Minute, zap.NewNop().Sugar())
	_, err := mgr.Submit(context.Background(), "test prompt", background.Origin{Channel: "cli"})
	require.NoError(t, err)
	time.Sleep(20 * time.Millisecond)

	cmd := NewBgCmd(func() (*background.Manager, error) { return mgr, nil })
	out, err := executeBgCommand(t, cmd, "list")

	require.NoError(t, err)
	assert.Contains(t, out, "STATUS")
	assert.Contains(t, out, "test prompt")
}

func TestBgStatus_WritesToCommandOutput(t *testing.T) {
	mgr := background.NewManager(stubRunner{result: "done"}, nil, 5, time.Minute, zap.NewNop().Sugar())
	id, err := mgr.Submit(context.Background(), "inspect me", background.Origin{Channel: "telegram", Session: "sess-1"})
	require.NoError(t, err)
	time.Sleep(20 * time.Millisecond)

	cmd := NewBgCmd(func() (*background.Manager, error) { return mgr, nil })
	out, err := executeBgCommand(t, cmd, "status", id)

	require.NoError(t, err)
	assert.Contains(t, out, "ID:")
	assert.Contains(t, out, "inspect me")
	assert.Contains(t, out, "telegram")
}

func TestBgResultAndCancel_WriteToCommandOutput(t *testing.T) {
	doneMgr := background.NewManager(stubRunner{result: "hello world"}, nil, 5, time.Minute, zap.NewNop().Sugar())
	doneID, err := doneMgr.Submit(context.Background(), "done prompt", background.Origin{})
	require.NoError(t, err)
	time.Sleep(120 * time.Millisecond)

	resultCmd := NewBgCmd(func() (*background.Manager, error) { return doneMgr, nil })
	resultOut, err := executeBgCommand(t, resultCmd, "result", doneID)
	require.NoError(t, err)
	assert.Contains(t, resultOut, "hello world")

	runningMgr := background.NewManager(stubRunner{delay: time.Second}, nil, 5, time.Minute, zap.NewNop().Sugar())
	cancelID, err := runningMgr.Submit(context.Background(), "cancel me", background.Origin{})
	require.NoError(t, err)
	time.Sleep(20 * time.Millisecond)

	cancelCmd := NewBgCmd(func() (*background.Manager, error) { return runningMgr, nil })
	cancelOut, err := executeBgCommand(t, cancelCmd, "cancel", cancelID)
	require.NoError(t, err)
	assert.Contains(t, cancelOut, "cancelled")
	assert.Contains(t, cancelOut, cancelID)
}
