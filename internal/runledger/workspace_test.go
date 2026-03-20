package runledger

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrepareStepWorkspace_RepeatedPreparationSucceeds(t *testing.T) {
	repoDir := t.TempDir()
	t.Chdir(repoDir)

	runCmd(t, repoDir, "git", "init")
	runCmd(t, repoDir, "git", "config", "user.email", "test@example.com")
	runCmd(t, repoDir, "git", "config", "user.name", "Test User")

	goFile := filepath.Join(repoDir, "main.go")
	require.NoError(t, os.WriteFile(goFile, []byte("package main\n"), 0o644))
	runCmd(t, repoDir, "git", "add", ".")
	runCmd(t, repoDir, "git", "-c", "commit.gpgsign=false", "commit", "-m", "init")

	ws := NewWorkspaceManager()
	step := &Step{
		StepID:    "s1",
		Goal:      "build",
		Validator: ValidatorSpec{Type: ValidatorBuildPass},
	}

	cleanup1, err := ws.PrepareStepWorkspace(step, "run-1")
	require.NoError(t, err)
	cleanup1()

	cleanup2, err := ws.PrepareStepWorkspace(step, "run-1")
	require.NoError(t, err)
	cleanup2()
}

func runCmd(t *testing.T, dir string, name string, args ...string) {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
}
