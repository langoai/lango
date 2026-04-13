package chat

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/langoai/lango/internal/background"
)

func TestTaskStripView_NilManager(t *testing.T) {
	m := taskStripModel{manager: nil}
	got := m.View(80)
	assert.Equal(t, "", got, "View with nil manager should return empty string")
}

func TestTaskStripView_NoSnapshots(t *testing.T) {
	m := taskStripModel{
		manager:   &background.Manager{},
		snapshots: nil,
	}
	got := m.View(80)
	assert.Equal(t, "", got, "View with no snapshots should return empty string")
}

func TestTaskStripView_SingleRunning(t *testing.T) {
	m := taskStripModel{
		manager: &background.Manager{},
		snapshots: []background.TaskSnapshot{
			{
				ID:         "t1",
				Status:     background.Running,
				StatusText: "running",
				Prompt:     "summarize the doc",
				StartedAt:  time.Now().Add(-5 * time.Second),
			},
		},
	}

	got := m.View(80)
	assert.NotEmpty(t, got, "View should produce output for a running task")
	assert.Contains(t, got, "summarize the doc", "should contain the task prompt")
	assert.Contains(t, got, "running", "should contain the running indicator")
}

func TestTaskStripView_LongPrompt(t *testing.T) {
	longPrompt := strings.Repeat("a", 50)
	m := taskStripModel{
		manager: &background.Manager{},
		snapshots: []background.TaskSnapshot{
			{
				ID:         "t2",
				Status:     background.Running,
				StatusText: "running",
				Prompt:     longPrompt,
				StartedAt:  time.Now(),
			},
		},
	}

	got := m.View(120)
	assert.NotEmpty(t, got)
	// The full 50-char prompt should NOT appear; it should be truncated to 27 + "..."
	assert.NotContains(t, got, longPrompt, "long prompt should be truncated")
	assert.Contains(t, got, "...", "truncated prompt should end with ellipsis")
}

func TestTaskStripView_NarrowWidth(t *testing.T) {
	m := taskStripModel{
		manager: &background.Manager{},
		snapshots: []background.TaskSnapshot{
			{
				ID:         "t3",
				Status:     background.Running,
				StatusText: "running",
				Prompt:     "short",
				StartedAt:  time.Now(),
			},
		},
	}

	// Should not panic with narrow width.
	got := m.View(20)
	assert.NotEmpty(t, got)
}

func TestTaskStripView_ZeroWidth(t *testing.T) {
	m := taskStripModel{
		manager: &background.Manager{},
		snapshots: []background.TaskSnapshot{
			{
				ID:         "t4",
				Status:     background.Running,
				StatusText: "running",
				Prompt:     "task",
				StartedAt:  time.Now(),
			},
		},
	}

	// Should not panic with zero width (max(0,1) guard).
	got := m.View(0)
	assert.NotEmpty(t, got)
}

func TestTaskStripView_CompletedTask(t *testing.T) {
	start := time.Now().Add(-10 * time.Second)
	completed := start.Add(3 * time.Second)

	m := taskStripModel{
		manager: &background.Manager{},
		snapshots: []background.TaskSnapshot{
			{
				ID:          "t5",
				Status:      background.Done,
				StatusText:  "done",
				Prompt:      "completed task",
				StartedAt:   start,
				CompletedAt: completed,
			},
		},
	}

	got := m.View(120)
	assert.NotEmpty(t, got)
	// Elapsed should be frozen at 3s, not growing with wall-clock.
	assert.Contains(t, got, "3s", "elapsed should be frozen at completed-started duration")
}

func TestTaskStripRefresh_NilManager(t *testing.T) {
	m := taskStripModel{manager: nil}
	m.refresh() // must not panic
	assert.Nil(t, m.snapshots, "snapshots should be nil after refresh with nil manager")
}
