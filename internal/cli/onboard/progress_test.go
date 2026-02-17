package onboard

import (
	"strings"
	"testing"
)

func TestRenderProgress_ContainsStepInfo(t *testing.T) {
	result := renderProgress(0, 80)

	if !strings.Contains(result, "[Step 1/5]") {
		t.Error("progress should contain step indicator [Step 1/5]")
	}
	if !strings.Contains(result, "Provider Setup") {
		t.Error("progress should contain step name")
	}
	// Should contain progress bar characters
	if !strings.Contains(result, "\u2501") {
		t.Error("progress should contain bar characters")
	}
}

func TestRenderProgress_AllSteps(t *testing.T) {
	for i := 0; i < len(WizardSteps); i++ {
		result := renderProgress(i, 80)
		expected := WizardSteps[i].Name
		if !strings.Contains(result, expected) {
			t.Errorf("step %d: want %q in output", i, expected)
		}
	}
}

func TestRenderStepList_CurrentHighlight(t *testing.T) {
	result := renderStepList(2)

	// Steps before current should have check mark
	if !strings.Contains(result, "\u2713") {
		t.Error("completed steps should have check mark")
	}
	// Current step should have pointer
	if !strings.Contains(result, "\u25b8") {
		t.Error("current step should have pointer indicator")
	}
	// Pending steps should have circle
	if !strings.Contains(result, "\u25cb") {
		t.Error("pending steps should have circle indicator")
	}
}

func TestRenderStepList_FirstStep(t *testing.T) {
	result := renderStepList(0)

	// No completed steps
	if strings.Contains(result, "\u2713") {
		t.Error("first step should have no completed indicators")
	}
	// Should have pointer for current
	if !strings.Contains(result, "\u25b8") {
		t.Error("first step should have pointer indicator")
	}
}

func TestRenderStepList_LastStep(t *testing.T) {
	result := renderStepList(len(WizardSteps) - 1)

	// Should have completed indicators for all previous steps
	count := strings.Count(result, "\u2713")
	if count != len(WizardSteps)-1 {
		t.Errorf("last step: want %d check marks, got %d", len(WizardSteps)-1, count)
	}
}
