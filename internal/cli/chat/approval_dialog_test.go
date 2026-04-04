package chat

import (
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
	t.Cleanup(func() {
		dialogScrollOffset = 0
		dialogSplitMode = false
	})

	vm := testVM()
	output := renderApprovalDialog(vm, 80, 40)

	assert.Contains(t, output, "HIGH", "should contain risk badge text")
	assert.Contains(t, output, "fs_edit", "should contain tool name")
	assert.Contains(t, output, "allow", "should contain action bar text")
}

func TestRenderApprovalDialog_NarrowWidth(t *testing.T) {
	t.Cleanup(func() {
		dialogScrollOffset = 0
		dialogSplitMode = false
	})

	vm := testVM()

	require.NotPanics(t, func() {
		output := renderApprovalDialog(vm, 30, 40)
		assert.NotEmpty(t, output, "output should be non-empty for narrow width")
	})
}

func TestRenderApprovalDialog_ShortHeight(t *testing.T) {
	t.Cleanup(func() {
		dialogScrollOffset = 0
		dialogSplitMode = false
	})

	vm := testVM()

	require.NotPanics(t, func() {
		output := renderApprovalDialog(vm, 80, 10)
		assert.NotEmpty(t, output, "output should be non-empty for short height")
	})
}

func TestRenderApprovalDialog_MinimalSize(t *testing.T) {
	t.Cleanup(func() {
		dialogScrollOffset = 0
		dialogSplitMode = false
	})

	vm := testVM()

	require.NotPanics(t, func() {
		output := renderApprovalDialog(vm, 10, 5)
		assert.NotEmpty(t, output, "output should be non-empty for minimal size")
	})
}

func TestRenderApprovalDialog_WithDiff(t *testing.T) {
	t.Cleanup(func() {
		dialogScrollOffset = 0
		dialogSplitMode = false
	})

	vm := testVM()
	vm.DiffContent = "+added line\n-removed line\n@@hunk header@@\nnormal line"

	output := renderApprovalDialog(vm, 80, 40)

	assert.Contains(t, output, "+added line", "should contain added diff line")
	assert.Contains(t, output, "-removed line", "should contain removed diff line")
	assert.Contains(t, output, "@@hunk header@@", "should contain hunk header")
}

func TestRenderApprovalDialog_DiffScroll(t *testing.T) {
	t.Cleanup(func() {
		dialogScrollOffset = 0
		dialogSplitMode = false
	})

	// Build a diff with many lines so scrolling changes visible content.
	var lines []string
	for i := 0; i < 50; i++ {
		lines = append(lines, "+line-"+strings.Repeat("x", 3)+"-"+string(rune('A'+i%26)))
	}
	vm := testVM()
	vm.DiffContent = strings.Join(lines, "\n")

	outputNoScroll := renderApprovalDialog(vm, 80, 40)

	dialogScrollOffset = 5
	outputScrolled := renderApprovalDialog(vm, 80, 40)

	assert.NotEqual(t, outputNoScroll, outputScrolled,
		"scrolled output should differ from non-scrolled output")
}

func TestRenderApprovalDialog_SplitMode(t *testing.T) {
	t.Cleanup(func() {
		dialogScrollOffset = 0
		dialogSplitMode = false
	})

	vm := testVM()
	vm.DiffContent = "+some change\n-old line"

	dialogSplitMode = true
	output := renderApprovalDialog(vm, 80, 40)

	assert.Contains(t, output, "split", "should contain split mode indicator in diff header")
}

func TestRenderApprovalDialog_EmptySummary(t *testing.T) {
	t.Cleanup(func() {
		dialogScrollOffset = 0
		dialogSplitMode = false
	})

	vm := testVM()
	vm.Request.Summary = ""

	output := renderApprovalDialog(vm, 80, 40)

	assert.Contains(t, output, "Execute tool:", "should contain fallback summary text")
}

func TestRenderApprovalDialog_WithParams(t *testing.T) {
	t.Cleanup(func() {
		dialogScrollOffset = 0
		dialogSplitMode = false
	})

	longValue := strings.Repeat("a", 200)
	vm := testVM()
	vm.Request.Params = map[string]interface{}{"content": longValue}

	output := renderApprovalDialog(vm, 80, 40)

	assert.Contains(t, output, "...", "long param values should be truncated with ellipsis")
	assert.NotContains(t, output, longValue, "full long value should not appear untruncated")
}

func TestRenderApprovalDialog_EmptyParams(t *testing.T) {
	t.Cleanup(func() {
		dialogScrollOffset = 0
		dialogSplitMode = false
	})

	vm := testVM()
	vm.Request.Params = nil

	output := renderApprovalDialog(vm, 80, 40)

	// With no params, the params section (key: value lines) should be absent.
	// Verify we still get the essentials without a params block.
	assert.Contains(t, output, "fs_edit", "should still contain tool name")
	assert.Contains(t, output, "allow", "should still contain action bar")
	// The param key "path" from default testVM should not appear since params is nil.
	assert.NotContains(t, output, "path:", "should not contain param key when params is empty")
}

// --- handleApprovalDialogKey tests ---

func TestHandleDialogKey_ScrollUp(t *testing.T) {
	t.Cleanup(func() {
		dialogScrollOffset = 0
		dialogSplitMode = false
	})

	dialogScrollOffset = 10
	upKey := tea.KeyMsg{Type: tea.KeyUp}
	cmd := handleApprovalDialogKey(upKey)

	assert.Nil(t, cmd, "scroll up should return nil cmd")
	assert.Equal(t, 7, dialogScrollOffset, "scroll up should decrease offset by 3")
}

func TestHandleDialogKey_ScrollDown(t *testing.T) {
	t.Cleanup(func() {
		dialogScrollOffset = 0
		dialogSplitMode = false
	})

	dialogScrollOffset = 0
	downKey := tea.KeyMsg{Type: tea.KeyDown}
	cmd := handleApprovalDialogKey(downKey)

	assert.Nil(t, cmd, "scroll down should return nil cmd")
	assert.Equal(t, 3, dialogScrollOffset, "scroll down should increase offset by 3")
}

func TestHandleDialogKey_ToggleSplit(t *testing.T) {
	t.Cleanup(func() {
		dialogScrollOffset = 0
		dialogSplitMode = false
	})

	dialogSplitMode = false
	tKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}

	cmd := handleApprovalDialogKey(tKey)
	assert.Nil(t, cmd, "toggle split should return nil cmd")
	assert.True(t, dialogSplitMode, "split mode should be true after toggle")

	cmd = handleApprovalDialogKey(tKey)
	assert.Nil(t, cmd, "toggle split should return nil cmd")
	assert.False(t, dialogSplitMode, "split mode should be false after second toggle")
}

func TestHandleDialogKey_Unhandled(t *testing.T) {
	t.Cleanup(func() {
		dialogScrollOffset = 0
		dialogSplitMode = false
	})

	dialogScrollOffset = 5
	dialogSplitMode = true

	xKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	cmd := handleApprovalDialogKey(xKey)

	assert.Nil(t, cmd, "unhandled key should return nil cmd")
	assert.Equal(t, 5, dialogScrollOffset, "scroll offset should not change for unhandled key")
	assert.True(t, dialogSplitMode, "split mode should not change for unhandled key")
}

// --- scrollApprovalDialog tests ---

func TestScrollDialog_NegativeClamp(t *testing.T) {
	t.Cleanup(func() {
		dialogScrollOffset = 0
		dialogSplitMode = false
	})

	dialogScrollOffset = 0
	scrollApprovalDialog(-5)

	assert.Equal(t, 0, dialogScrollOffset, "scroll offset should clamp to 0 when scrolling negative from 0")
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
