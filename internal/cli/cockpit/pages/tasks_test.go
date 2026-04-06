package pages

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/cli/tui"
)

// mockTaskLister implements TaskLister for testing.
type mockTaskLister struct {
	tasks []TaskInfo
}

func (m *mockTaskLister) ListTasks() []TaskInfo {
	return m.tasks
}

// mockTaskActioner implements TaskActioner for testing.
type mockTaskActioner struct {
	cancelCalled string
	retryCalled  string
	cancelErr    error
	retryErr     error
}

func (m *mockTaskActioner) CancelTask(id string) error {
	m.cancelCalled = id
	return m.cancelErr
}

func (m *mockTaskActioner) RetryTask(_ context.Context, id string) error {
	m.retryCalled = id
	return m.retryErr
}

// sampleTasks returns a set of three tasks for common test scenarios.
func sampleTasks() []TaskInfo {
	return []TaskInfo{
		{ID: "task-001", Prompt: "Summarize the document", Status: "running", Elapsed: 30 * time.Second},
		{ID: "task-002", Prompt: "Generate unit tests", Status: "pending", Elapsed: 10 * time.Second},
		{ID: "task-003", Prompt: "Refactor module", Status: "done", Elapsed: 2 * time.Minute},
	}
}

// sampleTasksWithAllStatuses returns tasks covering all action-relevant statuses.
func sampleTasksWithAllStatuses() []TaskInfo {
	return []TaskInfo{
		{ID: "t-run", Prompt: "Running task", Status: "running", Elapsed: 10 * time.Second},
		{ID: "t-pend", Prompt: "Pending task", Status: "pending", Elapsed: 5 * time.Second},
		{ID: "t-done", Prompt: "Done task", Status: "done", Elapsed: 60 * time.Second},
		{ID: "t-fail", Prompt: "Failed task", Status: "failed", Elapsed: 20 * time.Second, Error: "timeout"},
		{ID: "t-cancel", Prompt: "Cancelled task", Status: "cancelled", Elapsed: 15 * time.Second},
	}
}

// newTestTasksPage creates a TasksPage with the given lister and sets width/height via Update.
func newTestTasksPage(lister TaskLister, width, height int) *TasksPage {
	p := NewTasksPage(lister, nil)
	updated, _ := p.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return updated.(*TasksPage)
}

// newTestTasksPageWithActioner creates a TasksPage with a lister and actioner.
func newTestTasksPageWithActioner(lister TaskLister, actioner TaskActioner, width, height int) *TasksPage {
	p := NewTasksPage(lister, actioner)
	updated, _ := p.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return updated.(*TasksPage)
}

func TestTasksPage_Title(t *testing.T) {
	t.Parallel()

	p := NewTasksPage(nil, nil)
	assert.Equal(t, "Tasks", p.Title())
}

func TestTasksPage_ShortHelp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give       string
		actioner   TaskActioner
		detailMode bool
		wantCount  int
		wantEsc    bool
		wantCancel bool
	}{
		{
			give:      "nil actioner, no detail",
			wantCount: 3, // enter, up, down
		},
		{
			give:       "nil actioner, detail mode",
			detailMode: true,
			wantCount:  4, // + esc
			wantEsc:    true,
		},
		{
			give:       "with actioner",
			actioner:   &mockTaskActioner{},
			wantCount:  5, // enter, up, down, cancel, retry
			wantCancel: true,
		},
		{
			give:       "with actioner and detail",
			actioner:   &mockTaskActioner{},
			detailMode: true,
			wantCount:  6,
			wantEsc:    true,
			wantCancel: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			p := NewTasksPage(nil, tt.actioner)
			p.detailMode = tt.detailMode
			bindings := p.ShortHelp()
			assert.Len(t, bindings, tt.wantCount)
		})
	}
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
	p := NewTasksPage(lister, nil)

	cmd := p.Activate()
	assert.True(t, p.tickActive, "tickActive should be true after Activate")
	assert.NotNil(t, cmd, "Activate should return a tick command")
	assert.Len(t, p.tasks, 3, "tasks should be populated after Activate")
}

func TestTasksPage_Deactivate(t *testing.T) {
	t.Parallel()

	lister := &mockTaskLister{tasks: sampleTasks()}
	p := NewTasksPage(lister, nil)
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

// --- Detail mode tests ---

func TestTasksPage_DetailToggle(t *testing.T) {
	t.Parallel()

	lister := &mockTaskLister{tasks: sampleTasks()}
	p := newTestTasksPage(lister, 100, 40)
	p.Activate()

	assert.False(t, p.detailMode, "detail mode should be off initially")

	// Press enter to open detail.
	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	p = updated.(*TasksPage)
	assert.True(t, p.detailMode, "enter should open detail mode")
	assert.Equal(t, 0, p.detailScroll, "scroll should reset on open")

	// Press enter again to close detail.
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	p = updated.(*TasksPage)
	assert.False(t, p.detailMode, "enter should toggle detail mode off")
}

func TestTasksPage_DetailToggleEmptyTasks(t *testing.T) {
	t.Parallel()

	lister := &mockTaskLister{tasks: []TaskInfo{}}
	p := newTestTasksPage(lister, 100, 40)
	p.Activate()

	// Enter on empty list should not activate detail mode.
	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	p = updated.(*TasksPage)
	assert.False(t, p.detailMode)
}

func TestTasksPage_EscClosesDetail(t *testing.T) {
	t.Parallel()

	lister := &mockTaskLister{tasks: sampleTasks()}
	p := newTestTasksPage(lister, 100, 40)
	p.Activate()
	p.detailMode = true

	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyEsc})
	p = updated.(*TasksPage)
	assert.False(t, p.detailMode, "esc should close detail mode")
}

func TestTasksPage_DetailScroll(t *testing.T) {
	t.Parallel()

	lister := &mockTaskLister{tasks: sampleTasks()}
	p := newTestTasksPage(lister, 100, 40)
	p.Activate()
	p.detailMode = true

	// Down in detail mode scrolls detail, not cursor.
	initialCursor := p.cursor
	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyDown})
	p = updated.(*TasksPage)
	assert.Equal(t, 1, p.detailScroll, "down in detail mode should increase scroll")
	assert.Equal(t, initialCursor, p.cursor, "cursor should not move in detail mode")

	// Down again.
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyDown})
	p = updated.(*TasksPage)
	assert.Equal(t, 2, p.detailScroll)

	// Up decreases scroll.
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyUp})
	p = updated.(*TasksPage)
	assert.Equal(t, 1, p.detailScroll)

	// Up at 0 should not go negative.
	p.detailScroll = 0
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyUp})
	p = updated.(*TasksPage)
	assert.Equal(t, 0, p.detailScroll, "scroll should not go below 0")
}

func TestTasksPage_DetailViewRender(t *testing.T) {
	t.Parallel()

	lister := &mockTaskLister{tasks: []TaskInfo{
		{
			ID:            "task-detail",
			Prompt:        "Fix the payment bug",
			Status:        "running",
			Elapsed:       30 * time.Second,
			Result:        "Fixed the bug",
			Error:         "",
			OriginChannel: "telegram",
			TokensUsed:    1234,
		},
	}}
	p := newTestTasksPage(lister, 100, 40)
	p.Activate()
	p.detailMode = true

	view := p.View()
	assert.Contains(t, view, "Task Detail")
	assert.Contains(t, view, "running")
	assert.Contains(t, view, "30s")
	assert.Contains(t, view, "telegram")
	assert.Contains(t, view, "1,234")
	assert.Contains(t, view, "Fix the payment bug")
	assert.Contains(t, view, "Fixed the bug")
}

// --- Cancel/Retry action tests ---

func TestTasksPage_CancelRunningTask(t *testing.T) {
	t.Parallel()

	actioner := &mockTaskActioner{}
	lister := &mockTaskLister{tasks: sampleTasksWithAllStatuses()}
	p := newTestTasksPageWithActioner(lister, actioner, 100, 40)
	p.Activate()
	p.cursor = 0 // running task

	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	require.NotNil(t, cmd, "cancel on running task should produce a command")

	// Execute the command synchronously for testing.
	result := cmd()
	msg, ok := result.(taskActionResultMsg)
	require.True(t, ok)
	assert.Equal(t, "Cancelled: t-run", msg.msg)
	assert.NoError(t, msg.err)
	assert.Equal(t, "t-run", actioner.cancelCalled)
}

func TestTasksPage_CancelPendingTask(t *testing.T) {
	t.Parallel()

	actioner := &mockTaskActioner{}
	lister := &mockTaskLister{tasks: sampleTasksWithAllStatuses()}
	p := newTestTasksPageWithActioner(lister, actioner, 100, 40)
	p.Activate()
	p.cursor = 1 // pending task

	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	require.NotNil(t, cmd, "cancel on pending task should produce a command")

	result := cmd()
	msg, ok := result.(taskActionResultMsg)
	require.True(t, ok)
	assert.Equal(t, "Cancelled: t-pend", msg.msg)
	assert.Equal(t, "t-pend", actioner.cancelCalled)
}

func TestTasksPage_CancelIgnoredForDoneTask(t *testing.T) {
	t.Parallel()

	actioner := &mockTaskActioner{}
	lister := &mockTaskLister{tasks: sampleTasksWithAllStatuses()}
	p := newTestTasksPageWithActioner(lister, actioner, 100, 40)
	p.Activate()
	p.cursor = 2 // done task

	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	assert.Nil(t, cmd, "cancel on done task should be ignored")
	assert.Empty(t, actioner.cancelCalled)
}

func TestTasksPage_RetryFailedTask(t *testing.T) {
	t.Parallel()

	actioner := &mockTaskActioner{}
	lister := &mockTaskLister{tasks: sampleTasksWithAllStatuses()}
	p := newTestTasksPageWithActioner(lister, actioner, 100, 40)
	p.Activate()
	p.cursor = 3 // failed task

	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	require.NotNil(t, cmd, "retry on failed task should produce a command")

	result := cmd()
	msg, ok := result.(taskActionResultMsg)
	require.True(t, ok)
	assert.Equal(t, "Retried: t-fail", msg.msg)
	assert.NoError(t, msg.err)
	assert.Equal(t, "t-fail", actioner.retryCalled)
}

func TestTasksPage_RetryCancelledTask(t *testing.T) {
	t.Parallel()

	actioner := &mockTaskActioner{}
	lister := &mockTaskLister{tasks: sampleTasksWithAllStatuses()}
	p := newTestTasksPageWithActioner(lister, actioner, 100, 40)
	p.Activate()
	p.cursor = 4 // cancelled task

	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	require.NotNil(t, cmd, "retry on cancelled task should produce a command")

	result := cmd()
	msg, ok := result.(taskActionResultMsg)
	require.True(t, ok)
	assert.Equal(t, "Retried: t-cancel", msg.msg)
	assert.Equal(t, "t-cancel", actioner.retryCalled)
}

func TestTasksPage_RetryIgnoredForRunningTask(t *testing.T) {
	t.Parallel()

	actioner := &mockTaskActioner{}
	lister := &mockTaskLister{tasks: sampleTasksWithAllStatuses()}
	p := newTestTasksPageWithActioner(lister, actioner, 100, 40)
	p.Activate()
	p.cursor = 0 // running task

	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	assert.Nil(t, cmd, "retry on running task should be ignored")
	assert.Empty(t, actioner.retryCalled)
}

func TestTasksPage_NilActionerNoPanic(t *testing.T) {
	t.Parallel()

	lister := &mockTaskLister{tasks: sampleTasks()}
	p := newTestTasksPage(lister, 100, 40) // nil actioner
	p.Activate()

	// Press 'c' with nil actioner — should not panic.
	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	assert.Nil(t, cmd, "cancel with nil actioner should return nil cmd")

	// Press 'r' with nil actioner — should not panic.
	_, cmd = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	assert.Nil(t, cmd, "retry with nil actioner should return nil cmd")
}

func TestTasksPage_CancelErrorSetsStatusMsg(t *testing.T) {
	t.Parallel()

	actioner := &mockTaskActioner{cancelErr: errors.New("connection lost")}
	lister := &mockTaskLister{tasks: sampleTasksWithAllStatuses()}
	p := newTestTasksPageWithActioner(lister, actioner, 100, 40)
	p.Activate()
	p.cursor = 0 // running task

	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	require.NotNil(t, cmd)

	// Execute the command and feed the result message back.
	result := cmd()
	updated, _ := p.Update(result)
	p = updated.(*TasksPage)
	assert.Contains(t, p.statusMsg, "Error: connection lost")
}

func TestTasksPage_StatusMessageDisplay(t *testing.T) {
	t.Parallel()

	actioner := &mockTaskActioner{}
	lister := &mockTaskLister{tasks: sampleTasksWithAllStatuses()}
	p := newTestTasksPageWithActioner(lister, actioner, 100, 40)
	p.Activate()

	// Set a status message.
	p.statusMsg = "Cancelled: t-run"
	p.statusTime = time.Now()

	view := p.View()
	assert.Contains(t, view, "Cancelled: t-run")
}

func TestTasksPage_StatusMessageClearedAfterTTL(t *testing.T) {
	t.Parallel()

	lister := &mockTaskLister{tasks: sampleTasks()}
	p := newTestTasksPage(lister, 100, 40)
	p.Activate()

	// Set an expired status message.
	p.statusMsg = "old message"
	p.statusTime = time.Now().Add(-5 * time.Second) // 5s ago, past the 3s TTL

	// A tick should clear it.
	updated, _ := p.Update(taskTickMsg(time.Now()))
	p = updated.(*TasksPage)
	assert.Empty(t, p.statusMsg, "expired status message should be cleared on tick")
}

func TestTasksPage_ActionResultMsg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		msg       taskActionResultMsg
		wantMsg   string
	}{
		{
			give:    "success message",
			msg:     taskActionResultMsg{msg: "Cancelled: t1"},
			wantMsg: "Cancelled: t1",
		},
		{
			give:    "error message",
			msg:     taskActionResultMsg{err: errors.New("fail")},
			wantMsg: "Error: fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			p := newTestTasksPage(&mockTaskLister{tasks: sampleTasks()}, 100, 40)
			p.Activate()

			updated, _ := p.Update(tt.msg)
			p = updated.(*TasksPage)
			assert.Equal(t, tt.wantMsg, p.statusMsg)
			assert.False(t, p.statusTime.IsZero())
		})
	}
}

// --- Helper function tests ---

func TestFormatTokens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give int
		want string
	}{
		{give: 0, want: "0"},
		{give: 5, want: "5"},
		{give: 123, want: "123"},
		{give: 1234, want: "1,234"},
		{give: 12345, want: "12,345"},
		{give: 123456, want: "123,456"},
		{give: 1234567, want: "1,234,567"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tui.FormatTokens(tt.give))
		})
	}
}

func TestWordWrap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		giveWidth int
		want      string
	}{
		{give: "", giveWidth: 40, want: ""},
		{give: "short", giveWidth: 40, want: "short"},
		{give: "hello world foo bar", giveWidth: 12, want: "hello world\nfoo bar"},
		{give: "one two three", giveWidth: 5, want: "one\ntwo\nthree"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tui.WordWrap(tt.give, tt.giveWidth))
		})
	}
}
