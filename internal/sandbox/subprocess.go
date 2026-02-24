package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// SubprocessExecutor runs tool invocations in isolated child processes.
// The child process inherits only PATH and HOME environment variables,
// preventing access to in-memory secrets of the parent process.
type SubprocessExecutor struct {
	cfg Config
}

// NewSubprocessExecutor creates a subprocess executor with the given config.
func NewSubprocessExecutor(cfg Config) *SubprocessExecutor {
	return &SubprocessExecutor{cfg: cfg}
}

// Execute launches a child process running in sandbox worker mode and
// communicates via JSON over stdin/stdout.
func (e *SubprocessExecutor) Execute(ctx context.Context, toolName string, params map[string]interface{}) (map[string]interface{}, error) {
	// Apply per-tool timeout if configured.
	if e.cfg.TimeoutPerTool > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.cfg.TimeoutPerTool)
		defer cancel()
	}

	// Resolve the current executable path for the child process.
	selfPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("resolve executable path: %w", err)
	}

	cmd := exec.CommandContext(ctx, selfPath, workerFlag)

	// Clean environment: only PATH and HOME.
	cmd.Env = cleanEnv()

	// Prepare JSON request for stdin.
	req := ExecutionRequest{
		ToolName: toolName,
		Params:   params,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal execution request: %w", err)
	}
	cmd.Stdin = bytes.NewReader(reqBytes)

	// Capture stdout and stderr.
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the subprocess.
	if err := cmd.Run(); err != nil {
		// Check for timeout.
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("tool %q timed out after %v", toolName, e.cfg.TimeoutPerTool)
		}
		return nil, fmt.Errorf("subprocess execution of tool %q: %w (stderr: %s)", toolName, err, stderr.String())
	}

	// Parse result from stdout.
	var result ExecutionResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("unmarshal execution result: %w (raw: %s)", err, stdout.String())
	}

	if result.Error != "" {
		return nil, fmt.Errorf("tool %q: %s", toolName, result.Error)
	}

	return result.Output, nil
}

// cleanEnv returns a minimal environment with only PATH and HOME.
func cleanEnv() []string {
	var env []string
	if v := os.Getenv("PATH"); v != "" {
		env = append(env, "PATH="+v)
	}
	if v := os.Getenv("HOME"); v != "" {
		env = append(env, "HOME="+v)
	}
	return env
}
