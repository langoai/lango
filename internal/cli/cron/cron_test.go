package cron

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/testutil"
)

func executeCronCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestNewCronCmd_Structure(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	require.NotNil(t, cmd)
	assert.Equal(t, "cron", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestNewCronCmd_Subcommands(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	expected := []string{"add", "list", "delete", "pause", "resume", "history"}
	subCmds := make(map[string]bool, len(cmd.Commands()))
	for _, sub := range cmd.Commands() {
		subCmds[sub.Name()] = true
	}

	for _, name := range expected {
		assert.True(t, subCmds[name], "missing subcommand: %s", name)
	}
}

func TestListCmd_EmptyDB(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	out, err := executeCronCommand(t, cmd, "list")
	require.NoError(t, err)
	assert.Contains(t, out, "No cron jobs found.")
}

func TestListCmd_BootError(t *testing.T) {
	cmd := NewCronCmd(testutil.FailBootLoader(assert.AnError))

	_, err := executeCronCommand(t, cmd, "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bootstrap")
}

func TestHistoryCmd_EmptyDB(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	out, err := executeCronCommand(t, cmd, "history")
	require.NoError(t, err)
	assert.Contains(t, out, "No execution history found.")
}

func TestHistoryCmd_BootError(t *testing.T) {
	cmd := NewCronCmd(testutil.FailBootLoader(assert.AnError))

	_, err := executeCronCommand(t, cmd, "history")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bootstrap")
}

func TestAddCmd_MissingPrompt(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	_, err := executeCronCommand(t, cmd, "add", "--name", "test", "--schedule", "0 9 * * *")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--prompt is required")
}

func TestAddCmd_MissingName(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	_, err := executeCronCommand(t, cmd, "add", "--prompt", "do something", "--schedule", "0 9 * * *")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--name is required")
}

func TestAddCmd_MissingSchedule(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	_, err := executeCronCommand(t, cmd, "add", "--name", "test", "--prompt", "do something")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "one of --schedule, --every, or --at is required")
}

func TestAddCmd_MultipleSchedules(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	_, err := executeCronCommand(t, cmd, "add",
		"--name", "test",
		"--prompt", "do something",
		"--schedule", "0 9 * * *",
		"--every", "1h",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only one of --schedule, --every, or --at")
}

func TestAddCmd_HappyPath(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	out, err := executeCronCommand(t, cmd, "add",
		"--name", "test-job",
		"--prompt", "hello world",
		"--schedule", "0 9 * * *",
	)
	require.NoError(t, err)
	assert.Contains(t, out, `Cron job "test-job" created`)
	assert.Contains(t, out, "cron 0 9 * * *")
}

func TestAddCmd_WithEvery(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	out, err := executeCronCommand(t, cmd, "add",
		"--name", "interval-job",
		"--prompt", "check status",
		"--every", "30m",
	)
	require.NoError(t, err)
	assert.Contains(t, out, `Cron job "interval-job" created`)
	assert.Contains(t, out, "every 30m")
}

func TestDeleteCmd_MissingArg(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	_, err := executeCronCommand(t, cmd, "delete")
	require.Error(t, err)
}

func TestPauseCmd_MissingArg(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	_, err := executeCronCommand(t, cmd, "pause")
	require.Error(t, err)
}

func TestResumeCmd_MissingArg(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	_, err := executeCronCommand(t, cmd, "resume")
	require.Error(t, err)
}

func TestShortID(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{give: "abcdefgh-1234-5678", want: "abcdefgh"},
		{give: "short", want: "short"},
		{give: "12345678", want: "12345678"},
		{give: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, tt.want, shortID(tt.give))
		})
	}
}
