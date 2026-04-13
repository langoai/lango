package runledger

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalPayload(t *testing.T) {
	tests := []struct {
		give     interface{}
		wantJSON string
	}{
		{
			give:     RunCreatedPayload{SessionKey: "s1", Goal: "test goal"},
			wantJSON: `{"session_key":"s1","original_request":"","goal":"test goal"}`,
		},
		{
			give:     NoteWrittenPayload{Key: "k1", Value: "v1"},
			wantJSON: `{"key":"k1","value":"v1"}`,
		},
	}

	for _, tt := range tests {
		raw := marshalPayload(tt.give)
		assert.JSONEq(t, tt.wantJSON, string(raw))
	}
}

func TestMarshalPayload_LogsWarningOnError(t *testing.T) {
	// marshalPayload uses zap (logging.SubsystemSugar) which is a no-op in tests.
	// We verify the fallback behavior: unmarshalable input returns "{}".
	raw := marshalPayload(make(chan int))
	assert.Equal(t, "{}", string(raw))
}

func TestValidatorTypeConstants(t *testing.T) {
	types := []ValidatorType{
		ValidatorBuildPass,
		ValidatorTestPass,
		ValidatorFileChanged,
		ValidatorArtifactExists,
		ValidatorCommandPass,
		ValidatorOrchestratorApproval,
	}
	seen := make(map[ValidatorType]bool, len(types))
	for _, vt := range types {
		require.False(t, seen[vt], "duplicate validator type: %s", vt)
		seen[vt] = true
	}
}

func TestStepJSON(t *testing.T) {
	step := Step{
		StepID:     "step-1",
		Index:      0,
		Goal:       "implement feature",
		OwnerAgent: "operator",
		Status:     StepStatusPending,
		Validator: ValidatorSpec{
			Type:   ValidatorBuildPass,
			Target: "./...",
		},
		MaxRetries: DefaultMaxRetries,
	}

	data, err := json.Marshal(step)
	require.NoError(t, err)

	var decoded Step
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, step.StepID, decoded.StepID)
	assert.Equal(t, step.Goal, decoded.Goal)
	assert.Equal(t, StepStatusPending, decoded.Status)
	assert.Equal(t, ValidatorBuildPass, decoded.Validator.Type)
}
