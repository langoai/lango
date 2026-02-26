package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderPrompt_NoPlaceholders(t *testing.T) {
	rendered, err := RenderPrompt("Hello world", nil)
	require.NoError(t, err)
	assert.Equal(t, "Hello world", rendered)
}

func TestRenderPrompt_SingleSubstitution(t *testing.T) {
	results := map[string]string{"step1": "result-value"}
	rendered, err := RenderPrompt("Use {{step1.result}} here", results)
	require.NoError(t, err)
	assert.Equal(t, "Use result-value here", rendered)
}

func TestRenderPrompt_MultipleSubstitutions(t *testing.T) {
	results := map[string]string{
		"research": "research output",
		"analyze":  "analysis output",
	}
	tmpl := "Combine {{research.result}} with {{analyze.result}}"
	rendered, err := RenderPrompt(tmpl, results)
	require.NoError(t, err)
	assert.Equal(t, "Combine research output with analysis output", rendered)
}

func TestRenderPrompt_MissingKey(t *testing.T) {
	results := map[string]string{}
	_, err := RenderPrompt("Use {{missing.result}}", results)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing results for steps")
	assert.Contains(t, err.Error(), "missing")
}

func TestRenderPrompt_HyphenatedStepID(t *testing.T) {
	results := map[string]string{"my-step": "hyphen-value"}
	rendered, err := RenderPrompt("{{my-step.result}}", results)
	require.NoError(t, err)
	assert.Equal(t, "hyphen-value", rendered)
}

func TestRenderPrompt_UnderscoreStepID(t *testing.T) {
	results := map[string]string{"my_step": "underscore-value"}
	rendered, err := RenderPrompt("{{my_step.result}}", results)
	require.NoError(t, err)
	assert.Equal(t, "underscore-value", rendered)
}

func TestPlaceholderRe_Matches(t *testing.T) {
	tests := []struct {
		input   string
		matches bool
	}{
		{"{{step1.result}}", true},
		{"{{my-step.result}}", true},
		{"{{my_step.result}}", true},
		{"{{Step1.result}}", true},
		{"{{123.result}}", true},
		{"{{.result}}", false},          // empty step ID
		{"{{step1.output}}", false},     // wrong suffix
		{"{{ step1.result }}", false},   // spaces
		{"text without placeholders", false},
	}

	for _, tt := range tests {
		matched := placeholderRe.MatchString(tt.input)
		assert.Equal(t, tt.matches, matched, "input: %q", tt.input)
	}
}
