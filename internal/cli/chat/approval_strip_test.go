package chat

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"

	"github.com/langoai/lango/internal/approval"
)

func makeApprovalVM(toolName, summary string) approval.ApprovalViewModel {
	return approval.ApprovalViewModel{
		Request: approval.ApprovalRequest{
			ToolName: toolName,
			Summary:  summary,
		},
		Risk: approval.RiskIndicator{Level: "moderate", Label: "Reads file"},
	}
}

func TestRenderApprovalStrip_NormalWidth(t *testing.T) {
	vm := makeApprovalVM("fs_read", "Read config file")
	output := renderApprovalStrip(vm, 80)

	assert.Contains(t, output, "fs_read")
	assert.Contains(t, output, "[a]llow")
}

func TestRenderApprovalStrip_NarrowWidth(t *testing.T) {
	vm := makeApprovalVM("fs_read", "Read config file")
	output := renderApprovalStrip(vm, 30)

	assert.LessOrEqual(t, lipgloss.Width(output), 30)
}

func TestRenderApprovalStrip_VeryNarrow(t *testing.T) {
	vm := makeApprovalVM("fs_read", "Read a very important config file")

	assert.NotPanics(t, func() {
		renderApprovalStrip(vm, 1)
	})
}

func TestRenderApprovalStrip_LongSummary(t *testing.T) {
	longSummary := strings.Repeat("A long summary that should be truncated ", 3)
	vm := makeApprovalVM("fs_read", longSummary)
	output := renderApprovalStrip(vm, 60)

	// The output must fit within 60 columns.
	assert.LessOrEqual(t, lipgloss.Width(output), 60)
}

func TestRenderApprovalStrip_EmptySummary(t *testing.T) {
	vm := makeApprovalVM("fs_read", "")
	output := renderApprovalStrip(vm, 80)

	// Empty summary triggers fallback "Execute tool:" text.
	assert.Contains(t, output, "Execute tool:")
}

func TestRenderApprovalStrip_ZeroWidth(t *testing.T) {
	vm := makeApprovalVM("fs_read", "Read config file")

	// max(0,1)=1 guard should prevent panic.
	assert.NotPanics(t, func() {
		renderApprovalStrip(vm, 0)
	})
}

func TestRenderApprovalStrip_KoreanSummary(t *testing.T) {
	vm := makeApprovalVM("fs_read", "설정 파일을 읽습니다. 이 파일은 매우 중요한 설정을 포함하고 있습니다.")
	output := renderApprovalStrip(vm, 60)

	// Korean chars are double-width; output must still fit within 60.
	assert.LessOrEqual(t, lipgloss.Width(output), 60)
}

func TestRenderApprovalStrip_CriticalShowsDestructive(t *testing.T) {
	vm := makeApprovalVM("exec", "Run dangerous command")
	vm.Risk = approval.RiskIndicator{Level: "critical", Label: "Executes arbitrary code"}

	output := renderApprovalStrip(vm, 120)

	assert.Contains(t, output, "destructive", "critical risk should show destructive label")
}

func TestRenderApprovalStrip_NonCriticalNoDestructive(t *testing.T) {
	vm := makeApprovalVM("fs_read", "Read config file")
	vm.Risk = approval.RiskIndicator{Level: "moderate", Label: "Reads file"}

	output := renderApprovalStrip(vm, 120)

	assert.NotContains(t, output, "destructive", "non-critical risk should not show destructive label")
}

func TestRenderApprovalStrip_ConfirmPendingMessage(t *testing.T) {
	vm := makeApprovalVM("exec", "Run command")

	// Normal render.
	normalOutput := renderApprovalStrip(vm, 120)
	assert.Contains(t, normalOutput, "[a]llow", "normal should show allow key")
	assert.NotContains(t, normalOutput, "Press 'a' again", "normal should not show confirm prompt")

	// Confirm pending.
	confirmOutput := renderApprovalStrip(vm, 120, true)
	assert.Contains(t, confirmOutput, "Press 'a' again", "confirm pending should show re-press prompt")
}

func TestRenderApprovalStrip_SingleLineOutput(t *testing.T) {
	tests := []struct {
		give  string
		width int
	}{
		{give: "width 80", width: 80},
		{give: "width 40", width: 40},
		{give: "width 20", width: 20},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			vm := makeApprovalVM("fs_read", "Read config file")
			output := renderApprovalStrip(vm, tt.width)
			assert.Equal(t, 1, lipgloss.Height(output))
		})
	}
}
