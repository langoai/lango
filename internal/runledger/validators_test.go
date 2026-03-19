package runledger

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArtifactExistsValidator_WithWorkDir(t *testing.T) {
	dir := t.TempDir()
	// Create a file in the temp dir.
	err := os.WriteFile(filepath.Join(dir, "output.bin"), []byte("hello"), 0644)
	require.NoError(t, err)

	v := &ArtifactExistsValidator{}

	// With WorkDir: relative target resolved against WorkDir.
	result, err := v.Validate(context.Background(), ValidatorSpec{
		Type:    ValidatorArtifactExists,
		Target:  "output.bin",
		WorkDir: dir,
	}, nil)
	require.NoError(t, err)
	assert.True(t, result.Passed)

	// Without WorkDir: relative target resolved against cwd (unlikely to exist).
	result2, err := v.Validate(context.Background(), ValidatorSpec{
		Type:   ValidatorArtifactExists,
		Target: "nonexistent-artifact-12345.bin",
	}, nil)
	require.NoError(t, err)
	assert.False(t, result2.Passed)
}

func TestNeedsIsolation(t *testing.T) {
	tests := []struct {
		give ValidatorType
		want bool
	}{
		{ValidatorBuildPass, true},
		{ValidatorTestPass, true},
		{ValidatorFileChanged, true},
		{ValidatorArtifactExists, false},
		{ValidatorCommandPass, false},
		{ValidatorOrchestratorApproval, false},
	}

	for _, tt := range tests {
		step := &Step{Validator: ValidatorSpec{Type: tt.give}}
		assert.Equal(t, tt.want, NeedsIsolation(step), "validator type: %s", tt.give)
	}
}
