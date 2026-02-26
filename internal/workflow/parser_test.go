package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_ValidWorkflow(t *testing.T) {
	yaml := `
name: test-workflow
description: A test workflow
steps:
  - id: step1
    agent: executor
    prompt: "Do something"
  - id: step2
    agent: researcher
    prompt: "Research {{step1.result}}"
    depends_on: [step1]
`
	w, err := Parse([]byte(yaml))
	require.NoError(t, err)
	require.NotNil(t, w)
	assert.Equal(t, "test-workflow", w.Name)
	assert.Len(t, w.Steps, 2)
}

func TestParse_InvalidYAML(t *testing.T) {
	w, err := Parse([]byte("{{invalid yaml"))
	assert.Error(t, err)
	assert.Nil(t, w)
}

func TestValidate_EmptyName(t *testing.T) {
	w := &Workflow{Steps: []Step{{ID: "a"}}}
	err := Validate(w)
	assert.ErrorIs(t, err, ErrWorkflowNameEmpty)
}

func TestValidate_NoSteps(t *testing.T) {
	w := &Workflow{Name: "test"}
	err := Validate(w)
	assert.ErrorIs(t, err, ErrNoWorkflowSteps)
}

func TestValidate_EmptyStepID(t *testing.T) {
	w := &Workflow{
		Name:  "test",
		Steps: []Step{{ID: ""}},
	}
	err := Validate(w)
	assert.ErrorIs(t, err, ErrStepIDEmpty)
}

func TestValidate_DuplicateStepID(t *testing.T) {
	w := &Workflow{
		Name: "test",
		Steps: []Step{
			{ID: "a"},
			{ID: "a"},
		},
	}
	err := Validate(w)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate step id")
}

func TestValidate_UnknownDependency(t *testing.T) {
	w := &Workflow{
		Name: "test",
		Steps: []Step{
			{ID: "a", DependsOn: []string{"nonexistent"}},
		},
	}
	err := Validate(w)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown step")
}

func TestValidate_UnknownAgent(t *testing.T) {
	w := &Workflow{
		Name: "test",
		Steps: []Step{
			{ID: "a", Agent: "unknown-agent"},
		},
	}
	err := Validate(w)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown agent")
}

func TestValidate_ValidAgents(t *testing.T) {
	for _, agent := range []string{"executor", "researcher", "planner", "memory-manager"} {
		w := &Workflow{
			Name:  "test",
			Steps: []Step{{ID: "a", Agent: agent}},
		}
		assert.NoError(t, Validate(w), "agent %q should be valid", agent)
	}
}

func TestValidate_EmptyAgent_Valid(t *testing.T) {
	w := &Workflow{
		Name:  "test",
		Steps: []Step{{ID: "a", Agent: ""}},
	}
	assert.NoError(t, Validate(w), "empty agent should be valid (uses default)")
}

func TestValidate_CircularDependency(t *testing.T) {
	w := &Workflow{
		Name: "test",
		Steps: []Step{
			{ID: "a", DependsOn: []string{"c"}},
			{ID: "b", DependsOn: []string{"a"}},
			{ID: "c", DependsOn: []string{"b"}},
		},
	}
	err := Validate(w)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestDetectCycles_NoCycles(t *testing.T) {
	steps := []Step{
		{ID: "a"},
		{ID: "b", DependsOn: []string{"a"}},
		{ID: "c", DependsOn: []string{"a"}},
	}
	assert.NoError(t, detectCycles(steps))
}

func TestDetectCycles_DirectCycle(t *testing.T) {
	steps := []Step{
		{ID: "a", DependsOn: []string{"b"}},
		{ID: "b", DependsOn: []string{"a"}},
	}
	err := detectCycles(steps)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestDetectCycles_IndirectCycle(t *testing.T) {
	steps := []Step{
		{ID: "a", DependsOn: []string{"c"}},
		{ID: "b", DependsOn: []string{"a"}},
		{ID: "c", DependsOn: []string{"b"}},
	}
	err := detectCycles(steps)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}
