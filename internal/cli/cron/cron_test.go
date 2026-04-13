package cron

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/testutil"
)

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

	result := testutil.ExecCmdOK(t, cmd, "list")
	assert.Contains(t, result.Stdout, "No cron jobs found.")
}

func TestListCmd_BootError(t *testing.T) {
	cmd := NewCronCmd(testutil.FailBootLoader(assert.AnError))

	result := testutil.ExecCmd(t, cmd, "list")
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "bootstrap")
}

func TestHistoryCmd_EmptyDB(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	result := testutil.ExecCmdOK(t, cmd, "history")
	assert.Contains(t, result.Stdout, "No execution history found.")
}

func TestHistoryCmd_BootError(t *testing.T) {
	cmd := NewCronCmd(testutil.FailBootLoader(assert.AnError))

	result := testutil.ExecCmd(t, cmd, "history")
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "bootstrap")
}

func TestAddCmd_MissingPrompt(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	result := testutil.ExecCmd(t, cmd, "add", "--name", "test", "--schedule", "0 9 * * *")
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "--prompt is required")
}

func TestAddCmd_MissingName(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	result := testutil.ExecCmd(t, cmd, "add", "--prompt", "do something", "--schedule", "0 9 * * *")
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "--name is required")
}

func TestAddCmd_MissingSchedule(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	result := testutil.ExecCmd(t, cmd, "add", "--name", "test", "--prompt", "do something")
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "one of --schedule, --every, or --at is required")
}

func TestAddCmd_MultipleSchedules(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	result := testutil.ExecCmd(t, cmd, "add",
		"--name", "test",
		"--prompt", "do something",
		"--schedule", "0 9 * * *",
		"--every", "1h",
	)
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "only one of --schedule, --every, or --at")
}

func TestAddCmd_HappyPath(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	result := testutil.ExecCmdOK(t, cmd, "add",
		"--name", "test-job",
		"--prompt", "hello world",
		"--schedule", "0 9 * * *",
	)
	assert.Contains(t, result.Stdout, `Cron job "test-job" created`)
	assert.Contains(t, result.Stdout, "cron 0 9 * * *")
}

func TestAddCmd_WithEvery(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	result := testutil.ExecCmdOK(t, cmd, "add",
		"--name", "interval-job",
		"--prompt", "check status",
		"--every", "30m",
	)
	assert.Contains(t, result.Stdout, `Cron job "interval-job" created`)
	assert.Contains(t, result.Stdout, "every 30m")
}

func TestDeleteCmd_MissingArg(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	result := testutil.ExecCmd(t, cmd, "delete")
	require.Error(t, result.Err)
}

func TestPauseCmd_MissingArg(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	result := testutil.ExecCmd(t, cmd, "pause")
	require.Error(t, result.Err)
}

func TestResumeCmd_MissingArg(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewCronCmd(testutil.FakeBootLoader(t, cfg))

	result := testutil.ExecCmd(t, cmd, "resume")
	require.Error(t, result.Err)
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
