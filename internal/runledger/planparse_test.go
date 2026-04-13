package runledger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePlannerOutput_FencedJSON(t *testing.T) {
	raw := "Here is the plan:\n```json\n" + `{
  "goal": "implement feature X",
  "acceptance_criteria": [
    {"description": "build passes", "validator": {"type": "build_pass", "target": "./..."}}
  ],
  "steps": [
    {"id": "s1", "goal": "write code", "owner_agent": "operator", "validator": {"type": "build_pass"}}
  ]
}` + "\n```\n"

	plan, err := ParsePlannerOutput(raw)
	require.NoError(t, err)
	assert.Equal(t, "implement feature X", plan.Goal)
	assert.Len(t, plan.Steps, 1)
	assert.Len(t, plan.AcceptanceCriteria, 1)
}

func TestParsePlannerOutput_RawJSON(t *testing.T) {
	raw := `{"goal": "test", "steps": [{"id": "s1", "goal": "g", "owner_agent": "op", "validator": {"type": "build_pass"}}], "acceptance_criteria": []}`
	plan, err := ParsePlannerOutput(raw)
	require.NoError(t, err)
	assert.Equal(t, "test", plan.Goal)
}

func TestParsePlannerOutput_NoJSON(t *testing.T) {
	_, err := ParsePlannerOutput("no json here")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPlanJSON)
}

func TestParsePlannerOutput_InvalidJSON(t *testing.T) {
	raw := "```json\n{invalid}\n```"
	_, err := ParsePlannerOutput(raw)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPlanJSON)
}

func TestValidatePlanSchema_Valid(t *testing.T) {
	plan := &PlannerOutput{
		Goal: "build it",
		Steps: []StepInput{
			{ID: "s1", Goal: "write", OwnerAgent: "operator", Validator: ValidatorSpec{Type: ValidatorBuildPass}},
			{ID: "s2", Goal: "test", OwnerAgent: "operator", Validator: ValidatorSpec{Type: ValidatorTestPass}, DependsOn: []string{"s1"}},
		},
		AcceptanceCriteria: []AcceptanceCriteriaInput{
			{Description: "all good", Validator: ValidatorSpec{Type: ValidatorBuildPass}},
		},
	}

	err := ValidatePlanSchema(plan, []string{"operator", "navigator"})
	require.NoError(t, err)
}

func TestValidatePlanSchema_MissingGoal(t *testing.T) {
	plan := &PlannerOutput{
		Steps: []StepInput{
			{ID: "s1", Goal: "write", OwnerAgent: "op", Validator: ValidatorSpec{Type: ValidatorBuildPass}},
		},
	}
	err := ValidatePlanSchema(plan, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "goal is required")
}

func TestValidatePlanSchema_NoSteps(t *testing.T) {
	plan := &PlannerOutput{Goal: "test"}
	err := ValidatePlanSchema(plan, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one step")
}

func TestValidatePlanSchema_DuplicateID(t *testing.T) {
	plan := &PlannerOutput{
		Goal: "test",
		Steps: []StepInput{
			{ID: "s1", Goal: "a", OwnerAgent: "op", Validator: ValidatorSpec{Type: ValidatorBuildPass}},
			{ID: "s1", Goal: "b", OwnerAgent: "op", Validator: ValidatorSpec{Type: ValidatorBuildPass}},
		},
	}
	err := ValidatePlanSchema(plan, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate id")
}

func TestValidatePlanSchema_UnknownAgent(t *testing.T) {
	plan := &PlannerOutput{
		Goal: "test",
		Steps: []StepInput{
			{ID: "s1", Goal: "a", OwnerAgent: "unknown_agent", Validator: ValidatorSpec{Type: ValidatorBuildPass}},
		},
	}
	err := ValidatePlanSchema(plan, []string{"operator"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown agent")
}

func TestValidatePlanSchema_InvalidValidatorType(t *testing.T) {
	plan := &PlannerOutput{
		Goal: "test",
		Steps: []StepInput{
			{ID: "s1", Goal: "a", OwnerAgent: "op", Validator: ValidatorSpec{Type: "custom_magic"}},
		},
	}
	err := ValidatePlanSchema(plan, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown validator type")
}

func TestValidatePlanSchema_DependencyCycle(t *testing.T) {
	plan := &PlannerOutput{
		Goal: "test",
		Steps: []StepInput{
			{ID: "s1", Goal: "a", OwnerAgent: "op", Validator: ValidatorSpec{Type: ValidatorBuildPass}, DependsOn: []string{"s2"}},
			{ID: "s2", Goal: "b", OwnerAgent: "op", Validator: ValidatorSpec{Type: ValidatorBuildPass}, DependsOn: []string{"s1"}},
		},
	}
	err := ValidatePlanSchema(plan, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cycle")
}

func TestValidatePlanSchema_UnknownDependency(t *testing.T) {
	plan := &PlannerOutput{
		Goal: "test",
		Steps: []StepInput{
			{ID: "s1", Goal: "a", OwnerAgent: "op", Validator: ValidatorSpec{Type: ValidatorBuildPass}, DependsOn: []string{"nonexistent"}},
		},
	}
	err := ValidatePlanSchema(plan, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown step")
}

func TestValidatePlanSchema_MissingValidatorType(t *testing.T) {
	plan := &PlannerOutput{
		Goal: "test",
		Steps: []StepInput{
			{ID: "s1", Goal: "a", OwnerAgent: "op", Validator: ValidatorSpec{}},
		},
	}
	err := ValidatePlanSchema(plan, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validator type is required")
}

func TestConvertPlanToRunData(t *testing.T) {
	plan := &PlannerOutput{
		Goal: "build it",
		Steps: []StepInput{
			{ID: "s1", Goal: "write", OwnerAgent: "operator", Validator: ValidatorSpec{Type: ValidatorBuildPass}},
			{ID: "s2", Goal: "test", OwnerAgent: "operator", Validator: ValidatorSpec{Type: ValidatorTestPass}, DependsOn: []string{"s1"}},
		},
		AcceptanceCriteria: []AcceptanceCriteriaInput{
			{Description: "all good", Validator: ValidatorSpec{Type: ValidatorBuildPass}},
		},
	}

	steps, criteria := ConvertPlanToRunData(plan)
	assert.Len(t, steps, 2)
	assert.Equal(t, StepStatusPending, steps[0].Status)
	assert.Equal(t, DefaultMaxRetries, steps[0].MaxRetries)
	assert.Equal(t, []string{"coding"}, steps[0].ToolProfile)
	assert.Equal(t, []string{"s1"}, steps[1].DependsOn)
	assert.Len(t, criteria, 1)
	assert.False(t, criteria[0].Met)
}

func TestInferToolProfile(t *testing.T) {
	tests := []struct {
		give ValidatorType
		want []string
	}{
		{ValidatorBuildPass, []string{"coding"}},
		{ValidatorTestPass, []string{"coding"}},
		{ValidatorFileChanged, []string{"coding"}},
		{ValidatorOrchestratorApproval, []string{"supervisor"}},
		{ValidatorArtifactExists, []string{"coding"}},
		{ValidatorCommandPass, []string{"coding"}},
	}

	for _, tt := range tests {
		result := inferToolProfile(tt.give)
		assert.Equal(t, tt.want, result, "validator type: %s", tt.give)
	}
}
