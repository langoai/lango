package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/config"
)

func TestWave30CurrentWorkDirPrefersTrimmedConfigValue(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Tools.Exec.WorkDir = "  /tmp/lango-wave30  "

	assert.Equal(t, "/tmp/lango-wave30", currentWorkDir(cfg))
}

func TestWave30CurrentWorkDirFallsBackToProcessDirectory(t *testing.T) {
	t.Parallel()

	want, err := os.Getwd()
	require.NoError(t, err)

	assert.Equal(t, want, currentWorkDir(nil))
}

func TestWave30TaskElapsedHandlesPendingCompletedAndRunningSnapshots(t *testing.T) {
	t.Parallel()

	assert.Zero(t, taskElapsed(background.TaskSnapshot{}))

	startedAt := time.Date(2026, 5, 18, 10, 0, 0, 0, time.UTC)
	completed := background.TaskSnapshot{
		StartedAt:   startedAt,
		CompletedAt: startedAt.Add(1500 * time.Millisecond),
	}
	assert.Equal(t, 1500*time.Millisecond, taskElapsed(completed))

	running := background.TaskSnapshot{StartedAt: time.Now().Add(-20 * time.Millisecond)}
	assert.GreaterOrEqual(t, taskElapsed(running), 20*time.Millisecond)
}
