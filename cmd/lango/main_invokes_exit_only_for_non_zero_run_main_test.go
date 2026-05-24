package main

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/cli/chat"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/mcp"
)

func TestMainInvokesExitOnlyForNonZeroRunMain(t *testing.T) {
	origSandbox := isSandboxWorkerModeFn
	origBroker := isStorageBrokerModeFn
	origNewRoot := newRootCmdFn
	origExit := exitFn
	origStderr := mainStderr
	t.Cleanup(func() {
		isSandboxWorkerModeFn = origSandbox
		isStorageBrokerModeFn = origBroker
		newRootCmdFn = origNewRoot
		exitFn = origExit
		mainStderr = origStderr
	})

	isSandboxWorkerModeFn = func() bool { return false }
	isStorageBrokerModeFn = func() bool { return false }
	mainStderr = io.Discard

	var exitCodes []int
	exitFn = func(code int) {
		exitCodes = append(exitCodes, code)
	}

	newRootCmdFn = func() *cobra.Command {
		cmd := &cobra.Command{RunE: func(*cobra.Command, []string) error {
			return errors.New("mainInvokesExitOnlyForNonZeroRunMain6 root failure")
		}}
		cmd.SetArgs([]string{})
		cmd.SetErr(io.Discard)
		return cmd
	}
	main()
	assert.Equal(t, []int{1}, exitCodes)

	newRootCmdFn = func() *cobra.Command {
		cmd := &cobra.Command{Run: func(*cobra.Command, []string) {}}
		cmd.SetArgs([]string{})
		cmd.SetErr(io.Discard)
		return cmd
	}
	main()
	assert.Equal(t, []int{1}, exitCodes, "successful run must not call exitFn")
}

func TestRootCommandNonInteractiveShowsHelpWithoutWorkbench(t *testing.T) {
	origInteractive := isInteractiveFn
	origWorkbench := runWorkbenchFn
	t.Cleanup(func() {
		isInteractiveFn = origInteractive
		runWorkbenchFn = origWorkbench
	})

	isInteractiveFn = func() bool { return false }
	runWorkbenchFn = func(string) error {
		t.Fatal("workbench must not run when root command is non-interactive")
		return nil
	}

	var out bytes.Buffer
	cmd := newRootCmd()
	cmd.SetArgs([]string{})
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)

	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Lango is a high-performance AI agent")
	assert.Contains(t, out.String(), "Getting Started")
}

func TestRootCommandInteractivePassesModeToWorkbench(t *testing.T) {
	origInteractive := isInteractiveFn
	origWorkbench := runWorkbenchFn
	t.Cleanup(func() {
		isInteractiveFn = origInteractive
		runWorkbenchFn = origWorkbench
	})

	isInteractiveFn = func() bool { return true }

	var gotMode string
	runWorkbenchFn = func(mode string) error {
		gotMode = mode
		return nil
	}

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--mode", "research"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	require.NoError(t, cmd.Execute())
	assert.Equal(t, "research", gotMode)
}

func TestChatRuntimeFeaturesHandleNilAndMCPManager(t *testing.T) {
	t.Parallel()

	assert.Equal(t, chat.RuntimeFeatures{}, chatRuntimeFeatures(nil))
	assert.Equal(t, chat.RuntimeFeatures{}, chatRuntimeFeatures(&app.App{}))

	features := chatRuntimeFeatures(&app.App{
		MCPManager: mcp.NewServerManager(config.MCPConfig{}),
	})

	assert.True(t, features.MCPActive)
	assert.Equal(t, 0, features.MCPServerCount)
	assert.Equal(t, 0, features.MCPToolCount)
}
