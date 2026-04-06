package chat

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/approval"
)

func testVM() approval.ApprovalViewModel {
	return approval.ApprovalViewModel{
		Request: approval.ApprovalRequest{
			ToolName: "fs_edit",
			Summary:  "Edit config.yaml",
			Params:   map[string]interface{}{"path": "/etc/config.yaml"},
		},
		Risk: approval.RiskIndicator{Level: "high", Label: "Modifies filesystem"},
	}
}

// --- renderApprovalDialog tests ---

func TestRenderApprovalDialog_NormalSize(t *testing.T) {
	vm := testVM()
	state := &approvalState{}
	output := renderApprovalDialog(vm, state, 80, 40)

	assert.Contains(t, output, "HIGH", "should contain risk badge text")
	assert.Contains(t, output, "fs_edit", "should contain tool name")
	assert.Contains(t, output, "allow", "should contain action bar text")
}

func TestRenderApprovalDialog_NarrowWidth(t *testing.T) {
	vm := testVM()
	state := &approvalState{}

	require.NotPanics(t, func() {
		output := renderApprovalDialog(vm, state, 30, 40)
		assert.NotEmpty(t, output, "output should be non-empty for narrow width")
	})
}

func TestRenderApprovalDialog_ShortHeight(t *testing.T) {
	vm := testVM()
	state := &approvalState{}

	require.NotPanics(t, func() {
		output := renderApprovalDialog(vm, state, 80, 10)
		assert.NotEmpty(t, output, "output should be non-empty for short height")
	})
}

func TestRenderApprovalDialog_MinimalSize(t *testing.T) {
	vm := testVM()
	state := &approvalState{}

	require.NotPanics(t, func() {
		output := renderApprovalDialog(vm, state, 10, 5)
		assert.NotEmpty(t, output, "output should be non-empty for minimal size")
	})
}

func TestRenderApprovalDialog_WithDiff(t *testing.T) {
	vm := testVM()
	vm.DiffContent = "+added line\n-removed line\n@@hunk header@@\nnormal line"
	state := &approvalState{}

	output := renderApprovalDialog(vm, state, 80, 40)

	assert.Contains(t, output, "+added line", "should contain added diff line")
	assert.Contains(t, output, "-removed line", "should contain removed diff line")
	assert.Contains(t, output, "@@hunk header@@", "should contain hunk header")
}

func TestRenderApprovalDialog_DiffScroll(t *testing.T) {
	// Build a diff with many lines so scrolling changes visible content.
	var lines []string
	for i := 0; i < 50; i++ {
		lines = append(lines, "+line-"+strings.Repeat("x", 3)+"-"+string(rune('A'+i%26)))
	}
	vm := testVM()
	vm.DiffContent = strings.Join(lines, "\n")

	stateNoScroll := &approvalState{}
	outputNoScroll := renderApprovalDialog(vm, stateNoScroll, 80, 40)

	stateScrolled := &approvalState{scrollOffset: 5}
	outputScrolled := renderApprovalDialog(vm, stateScrolled, 80, 40)

	assert.NotEqual(t, outputNoScroll, outputScrolled,
		"scrolled output should differ from non-scrolled output")
}

func TestRenderApprovalDialog_SplitMode(t *testing.T) {
	vm := testVM()
	vm.DiffContent = "+some change\n-old line"
	state := &approvalState{splitMode: true}

	output := renderApprovalDialog(vm, state, 80, 40)

	assert.Contains(t, output, "split", "should contain split mode indicator in diff header")
}

func TestRenderApprovalDialog_EmptySummary(t *testing.T) {
	vm := testVM()
	vm.Request.Summary = ""
	state := &approvalState{}

	output := renderApprovalDialog(vm, state, 80, 40)

	assert.Contains(t, output, "Execute tool:", "should contain fallback summary text")
}

func TestRenderApprovalDialog_WithParams(t *testing.T) {
	longValue := strings.Repeat("a", 200)
	vm := testVM()
	vm.Request.Params = map[string]interface{}{"content": longValue}
	state := &approvalState{}

	output := renderApprovalDialog(vm, state, 80, 40)

	assert.Contains(t, output, "...", "long param values should be truncated with ellipsis")
	assert.NotContains(t, output, longValue, "full long value should not appear untruncated")
}

func TestRenderApprovalDialog_EmptyParams(t *testing.T) {
	vm := testVM()
	vm.Request.Params = nil
	state := &approvalState{}

	output := renderApprovalDialog(vm, state, 80, 40)

	// With no params, the params section (key: value lines) should be absent.
	// Verify we still get the essentials without a params block.
	assert.Contains(t, output, "fs_edit", "should still contain tool name")
	assert.Contains(t, output, "allow", "should still contain action bar")
	// The param key "path" from default testVM should not appear since params is nil.
	assert.NotContains(t, output, "path:", "should not contain param key when params is empty")
}

func TestRenderApprovalDialog_ShowsRuleExplanation(t *testing.T) {
	vm := testVM()
	vm.RuleExplanation = "This tool modifies the filesystem and is classified as dangerous."
	state := &approvalState{}

	output := renderApprovalDialog(vm, state, 80, 40)

	assert.Contains(t, output, "Why:", "should contain explanation prefix")
	assert.Contains(t, output, "filesystem", "should contain explanation text")
}

func TestRenderApprovalDialog_EmptyExplanationSkipped(t *testing.T) {
	vm := testVM()
	vm.RuleExplanation = ""
	state := &approvalState{}

	output := renderApprovalDialog(vm, state, 80, 40)

	assert.NotContains(t, output, "Why:", "should not contain explanation prefix when empty")
}

func TestRenderApprovalDialog_ConfirmPendingMessage(t *testing.T) {
	vm := testVM()

	// Normal render without confirmPending.
	stateNormal := &approvalState{}
	normalOutput := renderApprovalDialog(vm, stateNormal, 80, 40)
	assert.Contains(t, normalOutput, "allow", "normal should show allow action")
	assert.NotContains(t, normalOutput, "Press 'a' again", "normal should not show confirm prompt")

	// Render with confirmPending.
	stateConfirm := &approvalState{}
	confirmOutput := renderApprovalDialog(vm, stateConfirm, 80, 40, true)
	assert.Contains(t, confirmOutput, "Press 'a' again", "confirm pending should show re-press prompt")
}

// --- handleApprovalDialogKey tests ---

func TestHandleDialogKey_ScrollUp(t *testing.T) {
	state := &approvalState{scrollOffset: 10}
	upKey := tea.KeyMsg{Type: tea.KeyUp}
	cmd := handleApprovalDialogKey(upKey, state)

	assert.Nil(t, cmd, "scroll up should return nil cmd")
	assert.Equal(t, 7, state.scrollOffset, "scroll up should decrease offset by 3")
}

func TestHandleDialogKey_ScrollDown(t *testing.T) {
	state := &approvalState{scrollOffset: 0}
	downKey := tea.KeyMsg{Type: tea.KeyDown}
	cmd := handleApprovalDialogKey(downKey, state)

	assert.Nil(t, cmd, "scroll down should return nil cmd")
	assert.Equal(t, 3, state.scrollOffset, "scroll down should increase offset by 3")
}

func TestHandleDialogKey_ToggleSplit(t *testing.T) {
	state := &approvalState{splitMode: false}
	tKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}

	cmd := handleApprovalDialogKey(tKey, state)
	assert.Nil(t, cmd, "toggle split should return nil cmd")
	assert.True(t, state.splitMode, "split mode should be true after toggle")

	cmd = handleApprovalDialogKey(tKey, state)
	assert.Nil(t, cmd, "toggle split should return nil cmd")
	assert.False(t, state.splitMode, "split mode should be false after second toggle")
}

func TestHandleDialogKey_Unhandled(t *testing.T) {
	state := &approvalState{scrollOffset: 5, splitMode: true}

	xKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	cmd := handleApprovalDialogKey(xKey, state)

	assert.Nil(t, cmd, "unhandled key should return nil cmd")
	assert.Equal(t, 5, state.scrollOffset, "scroll offset should not change for unhandled key")
	assert.True(t, state.splitMode, "split mode should not change for unhandled key")
}

// --- diffLineCache tests ---

func TestDiffLineCache_HitOnSameContent(t *testing.T) {
	vm := testVM()
	vm.DiffContent = "+added\n-removed\nnormal"
	state := &approvalState{}

	// First render builds the cache.
	_ = renderApprovalDialog(vm, state, 80, 40)
	assert.NotNil(t, state.diffCache.lines, "cache should be populated after first render")
	assert.Equal(t, vm.DiffContent, state.diffCache.content, "cache content key should match")
	assert.Equal(t, 80, state.diffCache.width, "cache width key should match")

	cachedPtr := state.diffCache.lines

	// Second render with same params should reuse the cache.
	_ = renderApprovalDialog(vm, state, 80, 40)
	assert.Equal(t, cachedPtr, state.diffCache.lines, "cache should be reused on identical params")
}

func TestDiffLineCache_MissOnWidthChange(t *testing.T) {
	vm := testVM()
	vm.DiffContent = "+added\n-removed"
	state := &approvalState{}

	_ = renderApprovalDialog(vm, state, 80, 40)
	oldLines := state.diffCache.lines

	_ = renderApprovalDialog(vm, state, 100, 40)
	assert.Equal(t, 100, state.diffCache.width, "cache width should update on width change")
	assert.NotEqual(t, fmt.Sprintf("%p", oldLines), fmt.Sprintf("%p", state.diffCache.lines),
		"cache slice should be rebuilt on width change")
}

func TestDiffLineCache_InvalidatedByReset(t *testing.T) {
	state := &approvalState{}
	state.diffCache = diffLineCache{content: "x", width: 80, lines: []string{"cached"}}

	state.Reset(&ApprovalRequestMsg{})
	assert.Nil(t, state.diffCache.lines, "Reset should clear cache lines")
	assert.Equal(t, "", state.diffCache.content, "Reset should clear cache content")
}

func TestDiffLineCache_InvalidatedByToggleSplit(t *testing.T) {
	state := &approvalState{}
	state.diffCache = diffLineCache{content: "x", width: 80, lines: []string{"cached"}}

	state.ToggleSplit()
	assert.Nil(t, state.diffCache.lines, "ToggleSplit should clear cache lines")
}

// --- approvalState.ScrollDiff tests ---

func TestScrollDiff_NegativeClamp(t *testing.T) {
	state := &approvalState{scrollOffset: 0}
	state.ScrollDiff(-5)

	assert.Equal(t, 0, state.scrollOffset, "scroll offset should clamp to 0 when scrolling negative from 0")
}

// --- riskLevelColor tests ---

func TestRiskLevelColor_AllLevels(t *testing.T) {
	tests := []struct {
		give      string
		wantColor string
	}{
		{give: "critical", wantColor: string(riskLevelColor("critical"))},
		{give: "high", wantColor: string(riskLevelColor("high"))},
		{give: "moderate", wantColor: string(riskLevelColor("moderate"))},
		{give: "unknown", wantColor: string(riskLevelColor("unknown"))},
	}

	// Collect all colors to verify they are distinct.
	colors := make(map[string]string, len(tests))
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			color := riskLevelColor(tt.give)
			colorStr := string(color)
			assert.NotEmpty(t, colorStr, "color for %q should not be empty", tt.give)
			colors[tt.give] = colorStr
		})
	}

	// Verify all levels map to distinct colors.
	seen := make(map[string]string)
	for level, color := range colors {
		if prev, exists := seen[color]; exists {
			t.Errorf("risk levels %q and %q map to the same color %q", prev, level, color)
		}
		seen[color] = level
	}
}
