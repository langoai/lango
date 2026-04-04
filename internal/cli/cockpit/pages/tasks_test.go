package pages

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTaskLister implements TaskLister for testing.
type mockTaskLister struct {
	tasks []TaskInfo
}

func (m *mockTaskLister) ListTasks() []TaskInfo {
	return m.tasks
}

// sampleTasks returns a set of three tasks for common test scenarios.
func sampleTasks() []TaskInfo {
	return []TaskInfo{
		{ID: "task-001", Prompt: "Summarize the document", Status: "running", Elapsed: 30 * time.Second},
		{ID: "task-002", Prompt: "Generate unit tests", Status: "pending", Elapsed: 10 * time.Second},
		{ID: "task-003", Prompt: "Refactor module", Status: "done", Elapsed: 2 * time.Minute},
	}
}

// newTestTasksPage creates a TasksPage with the given lister and sets width/height via Update.
func newTestTasksPage(lister TaskLister, width, height int) *TasksPage {
	p := NewTasksPage(lister)
	updated, _ := p.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return updated.(*TasksPage)
}

func TestTasksPage_Title(t *testing.T) {
	t.Parallel()

	p := NewTasksPage(nil)
	assert.Equal(t, "Tasks", p.Title())
}

func TestTasksPage_ShortHelp(t *testing.T) {
	t.Parallel()

	p := NewTasksPage(nil)
	bindings := p.ShortHelp()
	assert.Nil(t, bindings)
}

func TestTasksPage_NilLister(t *testing.T) {
	t.Parallel()

	p := newTestTasksPage(nil, 80, 24)
	view := p.View()
	assert.Contains(t, view, "No background tasks")
}

func TestTasksPage_EmptyTasks(t *testing.T) {
	t.Parallel()

	lister := &mockTaskLister{tasks: []TaskInfo{}}
	p := newTestTasksPage(lister, 80, 24)
	p.Activate()
	view := p.View()
	assert.Contains(t, view, "No active tasks")
}

func TestTasksPage_WithTasks(t *testing.T) {
	t.Parallel()

	lister := &mockTaskLister{tasks: sampleTasks()}
	p := newTestTasksPage(lister, 100, 24)
	p.Activate()

	view := p.View()
	assert.Contains(t, view, "task-001")
	assert.Contains(t, view, "task-002")
	assert.Contains(t, view, "task-003")
	assert.Contains(t, view, "Summarize")
	assert.Contains(t, view, "Generate")
	assert.Contains(t, view, "Refactor")
}

func TestTasksPage_CursorDown(t *testing.T) {
	t.Parallel()

	lister := &mockTaskLister{tasks: sampleTasks()}
	p := newTestTasksPage(lister, 100, 24)
	p.Activate()

	assert.Equal(t, 0, p.cursor)

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyDown})
	p = updated.(*TasksPage)
	assert.Equal(t, 1, p.cursor)
}

func TestTasksPage_CursorUp(t *testing.T) {
	t.Parallel()

	lister := &mockTaskLister{tasks: sampleTasks()}
	p := newTestTasksPage(lister, 100, 24)
	p.Activate()
	p.cursor = 1

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyUp})
	p = updated.(*TasksPage)
	assert.Equal(t, 0, p.cursor)
}

func TestTasksPage_CursorClampBottom(t *testing.T) {
	t.Parallel()

	lister := &mockTaskLister{tasks: sampleTasks()}
	p := newTestTasksPage(lister, 100, 24)
	p.Activate()
	p.cursor = len(p.tasks) - 1 // last task

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyDown})
	p = updated.(*TasksPage)
	assert.Equal(t, len(p.tasks)-1, p.cursor, "cursor should not exceed last index")
}

func TestTasksPage_CursorClampTop(t *testing.T) {
	t.Parallel()

	lister := &mockTaskLister{tasks: sampleTasks()}
	p := newTestTasksPage(lister, 100, 24)
	p.Activate()
	assert.Equal(t, 0, p.cursor)

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyUp})
	p = updated.(*TasksPage)
	assert.Equal(t, 0, p.cursor, "cursor should not go below 0")
}

func TestTasksPage_CursorHighlight(t *testing.T) {
	t.Parallel()

	lister := &mockTaskLister{tasks: sampleTasks()}
	p := newTestTasksPage(lister, 100, 24)
	p.Activate()

	view := p.View()
	// The first row (cursor=0) should have the ">" prefix.
	assert.Contains(t, view, ">")

	// Move cursor to second row.
	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyDown})
	p = updated.(*TasksPage)
	view = p.View()

	// Split into lines and find lines containing task IDs.
	lines := strings.Split(view, "\n")
	var firstTaskLine, secondTaskLine string
	for _, line := range lines {
		if strings.Contains(line, "task-001") {
			firstTaskLine = line
		}
		if strings.Contains(line, "task-002") {
			secondTaskLine = line
		}
	}
	require.NotEmpty(t, firstTaskLine, "should find task-001 line")
	require.NotEmpty(t, secondTaskLine, "should find task-002 line")
	assert.NotContains(t, firstTaskLine, ">", "first task should not have cursor prefix")
	assert.Contains(t, secondTaskLine, ">", "second task should have cursor prefix")
}

func TestTasksPage_LongIDTruncated(t *testing.T) {
	t.Parallel()

	lister := &mockTaskLister{tasks: []TaskInfo{
		{ID: "abcdefghijkl", Prompt: "test prompt", Status: "running", Elapsed: 5 * time.Second},
	}}
	p := newTestTasksPage(lister, 100, 24)
	p.Activate()

	view := p.View()
	// The full 12-char ID should NOT appear; it is truncated to 8 chars (taskColIDW-2).
	assert.NotContains(t, view, "abcdefghijkl")
	assert.Contains(t, view, "abcde...")
}

func TestTasksPage_LongPromptTruncated(t *testing.T) {
	t.Parallel()

	longPrompt := strings.Repeat("x", 80)
	lister := &mockTaskLister{tasks: []TaskInfo{
		{ID: "t1", Prompt: longPrompt, Status: "running", Elapsed: 5 * time.Second},
	}}
	p := newTestTasksPage(lister, 80, 24)
	p.Activate()

	view := p.View()
	// The full 80-char prompt should not appear verbatim.
	assert.NotContains(t, view, longPrompt)
	// But a prefix should be present.
	assert.Contains(t, view, "xxxxx")
}

func TestTasksPage_Activate(t *testing.T) {
	t.Parallel()

	lister := &mockTaskLister{tasks: sampleTasks()}
	p := NewTasksPage(lister)

	cmd := p.Activate()
	assert.True(t, p.tickActive, "tickActive should be true after Activate")
	assert.NotNil(t, cmd, "Activate should return a tick command")
	assert.Len(t, p.tasks, 3, "tasks should be populated after Activate")
}

func TestTasksPage_Deactivate(t *testing.T) {
	t.Parallel()

	lister := &mockTaskLister{tasks: sampleTasks()}
	p := NewTasksPage(lister)
	p.Activate()
	assert.True(t, p.tickActive)

	p.Deactivate()
	assert.False(t, p.tickActive, "tickActive should be false after Deactivate")

	// A tick message after deactivation should not re-enable the tick.
	updated, cmd := p.Update(taskTickMsg(time.Now()))
	p = updated.(*TasksPage)
	assert.Nil(t, cmd, "tick command should not be returned when deactivated")
}

func TestTasksPage_NarrowWidth(t *testing.T) {
	t.Parallel()

	lister := &mockTaskLister{tasks: sampleTasks()}
	p := newTestTasksPage(lister, 30, 24)
	p.Activate()

	// Should not panic with narrow width.
	view := p.View()
	assert.NotEmpty(t, view)
	// Narrow mode should NOT show "Elapsed" header.
	assert.NotContains(t, view, "Elapsed")
	// Should still contain task IDs.
	assert.Contains(t, view, "task-001")
}

func TestTasksPage_WideWidth(t *testing.T) {
	t.Parallel()

	lister := &mockTaskLister{tasks: sampleTasks()}
	p := newTestTasksPage(lister, 100, 24)
	p.Activate()

	view := p.View()
	// Wide mode should show the "Elapsed" column header.
	assert.Contains(t, view, "Elapsed")
	// Should show elapsed values.
	assert.Contains(t, view, "30s")
	assert.Contains(t, view, "2m0s")
}
