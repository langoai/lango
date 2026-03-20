package runledger

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// BuildPassValidator verifies that `go build <target>` succeeds.
type BuildPassValidator struct{}

func (v *BuildPassValidator) Validate(ctx context.Context, spec ValidatorSpec, _ []Evidence) (*ValidationResult, error) {
	target := spec.Target
	if target == "" {
		target = "./..."
	}
	cmd := exec.CommandContext(ctx, "go", "build", target)
	if spec.WorkDir != "" {
		cmd.Dir = spec.WorkDir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ValidationResult{
			Passed: false,
			Reason: "build failed",
			Details: map[string]string{
				"exit_code": exitCodeStr(err),
				"output":    truncate(string(output), 2000),
			},
		}, nil
	}
	return &ValidationResult{Passed: true, Reason: "build succeeded"}, nil
}

// TestPassValidator verifies that `go test <target>` succeeds.
type TestPassValidator struct{}

func (v *TestPassValidator) Validate(ctx context.Context, spec ValidatorSpec, _ []Evidence) (*ValidationResult, error) {
	target := spec.Target
	if target == "" {
		target = "./..."
	}
	cmd := exec.CommandContext(ctx, "go", "test", target)
	if spec.WorkDir != "" {
		cmd.Dir = spec.WorkDir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ValidationResult{
			Passed: false,
			Reason: "tests failed",
			Details: map[string]string{
				"exit_code": exitCodeStr(err),
				"output":    truncate(string(output), 2000),
			},
		}, nil
	}
	return &ValidationResult{Passed: true, Reason: "all tests passed"}, nil
}

// FileChangedValidator verifies that files matching target pattern appear in git diff.
type FileChangedValidator struct{}

func (v *FileChangedValidator) Validate(ctx context.Context, spec ValidatorSpec, _ []Evidence) (*ValidationResult, error) {
	if spec.Target == "" {
		return &ValidationResult{
			Passed:  false,
			Reason:  "no target pattern specified",
			Missing: []string{"target"},
		}, nil
	}
	cmd := exec.CommandContext(ctx, "git", "diff", "--name-only", "HEAD")
	if spec.WorkDir != "" {
		cmd.Dir = spec.WorkDir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ValidationResult{
			Passed: false,
			Reason: "git diff failed",
			Details: map[string]string{
				"exit_code": exitCodeStr(err),
				"output":    truncate(string(output), 500),
			},
		}, nil
	}

	changedFiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	pattern := spec.Target
	var matched []string
	for _, f := range changedFiles {
		if f == "" {
			continue
		}
		ok, _ := filepath.Match(pattern, f)
		if ok || strings.Contains(f, pattern) {
			matched = append(matched, f)
		}
	}
	if len(matched) == 0 {
		return &ValidationResult{
			Passed:  false,
			Reason:  fmt.Sprintf("no changed files match pattern %q", pattern),
			Missing: []string{pattern},
		}, nil
	}
	return &ValidationResult{
		Passed: true,
		Reason: fmt.Sprintf("%d file(s) changed matching %q", len(matched), pattern),
		Details: map[string]string{
			"matched_files": strings.Join(matched, ", "),
		},
	}, nil
}

// ArtifactExistsValidator verifies that a file exists at the target path.
type ArtifactExistsValidator struct{}

func (v *ArtifactExistsValidator) Validate(_ context.Context, spec ValidatorSpec, _ []Evidence) (*ValidationResult, error) {
	if spec.Target == "" {
		return &ValidationResult{
			Passed:  false,
			Reason:  "no target path specified",
			Missing: []string{"target"},
		}, nil
	}
	target := spec.Target
	if spec.WorkDir != "" {
		target = filepath.Join(spec.WorkDir, spec.Target)
	}
	if _, err := os.Stat(target); err != nil {
		return &ValidationResult{
			Passed:  false,
			Reason:  fmt.Sprintf("artifact not found: %s", spec.Target),
			Missing: []string{spec.Target},
		}, nil
	}
	return &ValidationResult{
		Passed: true,
		Reason: fmt.Sprintf("artifact exists: %s", spec.Target),
	}, nil
}

// CommandPassValidator runs an arbitrary command and checks the exit code.
type CommandPassValidator struct{}

func (v *CommandPassValidator) Validate(ctx context.Context, spec ValidatorSpec, _ []Evidence) (*ValidationResult, error) {
	if spec.Target == "" {
		return &ValidationResult{
			Passed:  false,
			Reason:  "no command specified",
			Missing: []string{"target"},
		}, nil
	}

	expectedExit := 0
	if e, ok := spec.Params["expected_exit_code"]; ok {
		if n, err := strconv.Atoi(e); err == nil {
			expectedExit = n
		}
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", spec.Target)
	if spec.WorkDir != "" {
		cmd.Dir = spec.WorkDir
	}
	output, err := cmd.CombinedOutput()

	actualExit := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			actualExit = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("run command: %w", err)
		}
	}

	if actualExit == expectedExit {
		return &ValidationResult{
			Passed: true,
			Reason: fmt.Sprintf("command exited with expected code %d", expectedExit),
			Details: map[string]string{
				"exit_code": strconv.Itoa(actualExit),
				"stdout":    truncate(string(output), 1000),
			},
		}, nil
	}

	return &ValidationResult{
		Passed: false,
		Reason: fmt.Sprintf("command exited with %d (expected %d)", actualExit, expectedExit),
		Details: map[string]string{
			"exit_code":          strconv.Itoa(actualExit),
			"expected_exit_code": strconv.Itoa(expectedExit),
			"output":             truncate(string(output), 2000),
		},
	}, nil
}

// OrchestratorApprovalValidator never auto-passes.
// It returns a failed result that triggers a PolicyRequest,
// which the orchestrator must explicitly approve via run_approve_step.
type OrchestratorApprovalValidator struct{}

func (v *OrchestratorApprovalValidator) Validate(_ context.Context, _ ValidatorSpec, _ []Evidence) (*ValidationResult, error) {
	return &ValidationResult{
		Passed: false,
		Reason: "awaiting orchestrator approval",
	}, nil
}

// DefaultValidators returns the standard validator set.
func DefaultValidators() map[ValidatorType]Validator {
	return map[ValidatorType]Validator{
		ValidatorBuildPass:            &BuildPassValidator{},
		ValidatorTestPass:             &TestPassValidator{},
		ValidatorFileChanged:          &FileChangedValidator{},
		ValidatorArtifactExists:       &ArtifactExistsValidator{},
		ValidatorCommandPass:          &CommandPassValidator{},
		ValidatorOrchestratorApproval: &OrchestratorApprovalValidator{},
	}
}

func exitCodeStr(err error) string {
	if exitErr, ok := err.(*exec.ExitError); ok {
		return strconv.Itoa(exitErr.ExitCode())
	}
	return "-1"
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "...(truncated)"
}
