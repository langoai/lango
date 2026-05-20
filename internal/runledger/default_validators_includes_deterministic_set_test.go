package runledger

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultValidatorsIncludesDeterministicSet(t *testing.T) {
	t.Parallel()

	validators := DefaultValidators()

	for _, validatorType := range []ValidatorType{
		ValidatorBuildPass,
		ValidatorTestPass,
		ValidatorFileChanged,
		ValidatorArtifactExists,
		ValidatorCommandPass,
		ValidatorOrchestratorApproval,
	} {
		assert.NotNil(t, validators[validatorType], "validator %s should be registered", validatorType)
	}
}

func TestArtifactExistsValidatorValidatesMissingTargetAndPath(t *testing.T) {
	t.Parallel()

	validator := &ArtifactExistsValidator{}

	result, err := validator.Validate(context.Background(), ValidatorSpec{}, nil)
	require.NoError(t, err)
	assert.False(t, result.Passed)
	assert.Equal(t, "no target path specified", result.Reason)
	assert.Equal(t, []string{"target"}, result.Missing)

	result, err = validator.Validate(context.Background(), ValidatorSpec{
		Target:  "missing.txt",
		WorkDir: t.TempDir(),
	}, nil)
	require.NoError(t, err)
	assert.False(t, result.Passed)
	assert.Equal(t, "artifact not found: missing.txt", result.Reason)
	assert.Equal(t, []string{"missing.txt"}, result.Missing)
}

func TestFileChangedValidatorValidationAndGitError(t *testing.T) {
	t.Parallel()

	validator := &FileChangedValidator{}

	result, err := validator.Validate(context.Background(), ValidatorSpec{}, nil)
	require.NoError(t, err)
	assert.False(t, result.Passed)
	assert.Equal(t, "no target pattern specified", result.Reason)
	assert.Equal(t, []string{"target"}, result.Missing)

	result, err = validator.Validate(context.Background(), ValidatorSpec{
		Target:  "*.go",
		WorkDir: filepath.Join(t.TempDir(), "does-not-exist"),
	}, nil)
	require.NoError(t, err)
	assert.False(t, result.Passed)
	assert.Equal(t, "git diff failed", result.Reason)
	assert.Equal(t, "-1", result.Details["exit_code"])
}

func TestFileChangedValidatorMatchesDiffByGlobOrSubstring(t *testing.T) {
	t.Parallel()

	requireDefaultValidatorsIncludesDeterministicSetCommand(t, "git")
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "initSecurityDeterministicDisabledAndErrorBranches4@example.test")
	runGit(t, dir, "config", "user.name", "Coverage Validator")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "internal", "runledger"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "internal", "runledger", "validators.go"), []byte("package runledger\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "internal", "runledger", "validators_test.go"), []byte("package runledger\n"), 0644))
	runGit(t, dir, "add", ".")
	runGit(t, dir, "-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null", "commit", "-m", "baseline")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "internal", "runledger", "validators_test.go"), []byte("package runledger\n\n// changed\n"), 0644))

	validator := &FileChangedValidator{}
	result, err := validator.Validate(defaultValidatorsIncludesDeterministicSetValidatorContext(t), ValidatorSpec{
		Target:  "internal/runledger/*_test.go",
		WorkDir: dir,
	}, nil)
	require.NoError(t, err)
	require.True(t, result.Passed)
	assert.Contains(t, result.Reason, "1 file(s) changed")
	assert.Equal(t, "internal/runledger/validators_test.go", result.Details["matched_files"])

	result, err = validator.Validate(defaultValidatorsIncludesDeterministicSetValidatorContext(t), ValidatorSpec{
		Target:  "validators_test",
		WorkDir: dir,
	}, nil)
	require.NoError(t, err)
	require.True(t, result.Passed)
	assert.Equal(t, "internal/runledger/validators_test.go", result.Details["matched_files"])
}

func TestCommandPassValidatorCoversExpectedExitAndErrors(t *testing.T) {
	t.Parallel()

	requireDefaultValidatorsIncludesDeterministicSetCommand(t, "sh")
	validator := &CommandPassValidator{}

	result, err := validator.Validate(context.Background(), ValidatorSpec{}, nil)
	require.NoError(t, err)
	assert.False(t, result.Passed)
	assert.Equal(t, "no command specified", result.Reason)

	result, err = validator.Validate(defaultValidatorsIncludesDeterministicSetValidatorContext(t), ValidatorSpec{
		Target: "printf 'expected failure'; exit 7",
		Params: map[string]string{"expected_exit_code": "7"},
	}, nil)
	require.NoError(t, err)
	require.True(t, result.Passed)
	assert.Equal(t, "command exited with expected code 7", result.Reason)
	assert.Equal(t, "7", result.Details["exit_code"])
	assert.Equal(t, "expected failure", result.Details["stdout"])

	result, err = validator.Validate(defaultValidatorsIncludesDeterministicSetValidatorContext(t), ValidatorSpec{
		Target: "printf 'nope'; exit 3",
		Params: map[string]string{"expected_exit_code": "not-a-number"},
	}, nil)
	require.NoError(t, err)
	require.False(t, result.Passed)
	assert.Equal(t, "command exited with 3 (expected 0)", result.Reason)
	assert.Equal(t, "3", result.Details["exit_code"])
	assert.Equal(t, "0", result.Details["expected_exit_code"])
	assert.Equal(t, "nope", result.Details["output"])

	result, err = validator.Validate(defaultValidatorsIncludesDeterministicSetValidatorContext(t), ValidatorSpec{
		Target:  "printf 'unreachable'",
		WorkDir: filepath.Join(t.TempDir(), "missing"),
	}, nil)
	assert.Nil(t, result)
	assert.ErrorContains(t, err, "run command")
}

func TestBuildAndTestValidatorsReturnStructuredFailures(t *testing.T) {
	t.Parallel()

	requireDefaultValidatorsIncludesDeterministicSetCommand(t, "go")
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module initSecurityDeterministicDisabledAndErrorBranches4.example\n\ngo 1.24\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "broken.go"), []byte("package initSecurityDeterministicDisabledAndErrorBranches4\n\nfunc Broken(\n"), 0644))

	buildResult, err := (&BuildPassValidator{}).Validate(defaultValidatorsIncludesDeterministicSetValidatorContext(t), ValidatorSpec{Target: ".", WorkDir: dir}, nil)
	require.NoError(t, err)
	require.False(t, buildResult.Passed)
	assert.Equal(t, "build failed", buildResult.Reason)
	assert.NotEmpty(t, buildResult.Details["exit_code"])
	assert.Contains(t, buildResult.Details["output"], "syntax error")

	require.NoError(t, os.WriteFile(filepath.Join(dir, "broken.go"), []byte("package initSecurityDeterministicDisabledAndErrorBranches4\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "broken_test.go"), []byte("package initSecurityDeterministicDisabledAndErrorBranches4\n\nimport \"testing\"\n\nfunc TestBroken(t *testing.T) { t.Fatal(\"boom\") }\n"), 0644))

	testResult, err := (&TestPassValidator{}).Validate(defaultValidatorsIncludesDeterministicSetValidatorContext(t), ValidatorSpec{Target: ".", WorkDir: dir}, nil)
	require.NoError(t, err)
	require.False(t, testResult.Passed)
	assert.Equal(t, "tests failed", testResult.Reason)
	assert.NotEmpty(t, testResult.Details["exit_code"])
	assert.Contains(t, testResult.Details["output"], "boom")
}

func TestBuildAndTestValidatorsUseDefaultTarget(t *testing.T) {
	t.Parallel()

	requireDefaultValidatorsIncludesDeterministicSetCommand(t, "go")
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module initSecurityDeterministicDisabledAndErrorBranches4.example\n\ngo 1.24\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ok.go"), []byte("package initSecurityDeterministicDisabledAndErrorBranches4\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ok_test.go"), []byte("package initSecurityDeterministicDisabledAndErrorBranches4\n\nimport \"testing\"\n\nfunc TestOK(t *testing.T) {}\n"), 0644))

	buildResult, err := (&BuildPassValidator{}).Validate(defaultValidatorsIncludesDeterministicSetValidatorContext(t), ValidatorSpec{WorkDir: dir}, nil)
	require.NoError(t, err)
	assert.True(t, buildResult.Passed)
	assert.Equal(t, "build succeeded", buildResult.Reason)

	testResult, err := (&TestPassValidator{}).Validate(defaultValidatorsIncludesDeterministicSetValidatorContext(t), ValidatorSpec{WorkDir: dir}, nil)
	require.NoError(t, err)
	assert.True(t, testResult.Passed)
	assert.Equal(t, "all tests passed", testResult.Reason)
}

func TestOrchestratorApprovalValidatorAlwaysRequiresApproval(t *testing.T) {
	t.Parallel()

	result, err := (&OrchestratorApprovalValidator{}).Validate(context.Background(), ValidatorSpec{}, nil)
	require.NoError(t, err)
	assert.False(t, result.Passed)
	assert.Equal(t, "awaiting orchestrator approval", result.Reason)
}

func TestExitCodeStrAndTruncate(t *testing.T) {
	t.Parallel()

	requireDefaultValidatorsIncludesDeterministicSetCommand(t, "sh")
	ctx := defaultValidatorsIncludesDeterministicSetValidatorContext(t)
	err := exec.CommandContext(ctx, "sh", "-c", "exit 42").Run()
	require.Error(t, err)
	assert.Equal(t, "42", exitCodeStr(err))
	assert.Equal(t, "-1", exitCodeStr(errors.New("plain error")))

	assert.Equal(t, "short", truncate("short", 10))
	truncated := truncate(strings.Repeat("x", 12), 5)
	assert.Equal(t, "xxxxx...(truncated)", truncated)
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	ctx := defaultValidatorsIncludesDeterministicSetValidatorContext(t)
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, output)
}

func requireDefaultValidatorsIncludesDeterministicSetCommand(t *testing.T, name string) {
	t.Helper()

	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s is required for this validator test: %v", name, err)
	}
}

func defaultValidatorsIncludesDeterministicSetValidatorContext(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)
	return ctx
}
