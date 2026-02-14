package exec

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/langowarny/lango/internal/logging"
	"github.com/langowarny/lango/internal/security"
)

var logger = logging.SubsystemSugar("tool.exec")

// Config holds exec tool configuration
type Config struct {
	DefaultTimeout  time.Duration
	AllowBackground bool
	WorkDir         string
	EnvFilter       []string // environment variables to exclude
	EnvWhitelist    []string // if set, ONLY these vars are allowed
	Refs            *security.RefStore // secret reference token resolver
}

// Tool provides shell command execution
type Tool struct {
	config      Config
	bgProcesses map[string]*BackgroundProcess
	bgMu        sync.RWMutex
}

// BackgroundProcess represents a running background command
type BackgroundProcess struct {
	ID        string
	Command   string
	Cmd       *exec.Cmd
	Output    *bytes.Buffer
	StartTime time.Time
	Done      bool
	ExitCode  int
	Error     string
}

// Result represents command execution result
type Result struct {
	ExitCode int    `json:"exitCode"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	TimedOut bool   `json:"timedOut,omitempty"`
}

// New creates a new exec tool
func New(cfg Config) *Tool {
	if cfg.DefaultTimeout == 0 {
		cfg.DefaultTimeout = 30 * time.Second
	}
	return &Tool{
		config:      cfg,
		bgProcesses: make(map[string]*BackgroundProcess),
	}
}

// resolveRefs resolves any secret reference tokens in the command string.
// Tokens like {{secret:name}} and {{decrypt:id}} are replaced with actual values
// just before execution. The resolved command is never logged or returned to the agent.
func (t *Tool) resolveRefs(command string) string {
	if t.config.Refs == nil {
		return command
	}
	return t.config.Refs.ResolveAll(command)
}

// Run executes a command synchronously
func (t *Tool) Run(ctx context.Context, command string, timeout time.Duration) (*Result, error) {
	if timeout == 0 {
		timeout = t.config.DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Resolve secret reference tokens just before execution.
	// The resolved command is never logged or sent back to the agent.
	resolved := t.resolveRefs(command)
	cmd := exec.CommandContext(ctx, "sh", "-c", resolved)
	cmd.Dir = t.config.WorkDir
	cmd.Env = t.filterEnv(os.Environ())

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	logger.Infow("executing command", "command", command, "timeout", timeout)

	err := cmd.Run()

	result := &Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		result.ExitCode = -1
		return result, nil
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return nil, err
		}
	}

	return result, nil
}

// RunWithPTY executes a command with PTY support
func (t *Tool) RunWithPTY(ctx context.Context, command string, timeout time.Duration) (*Result, error) {
	if timeout == 0 {
		timeout = t.config.DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resolved := t.resolveRefs(command)
	cmd := exec.CommandContext(ctx, "sh", "-c", resolved)
	cmd.Dir = t.config.WorkDir
	cmd.Env = t.filterEnv(os.Environ())

	// Start with PTY
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to start PTY: %w", err)
	}
	defer ptmx.Close()

	// Read output
	var output bytes.Buffer
	done := make(chan error, 1)

	go func() {
		_, err := io.Copy(&output, ptmx)
		done <- err
	}()

	// Wait for completion or timeout
	select {
	case <-ctx.Done():
		cmd.Process.Signal(syscall.SIGTERM)
		return &Result{
			Stdout:   output.String(),
			TimedOut: true,
			ExitCode: -1,
		}, nil
	case <-done:
		// Process completed
	}

	// Wait for process
	err = cmd.Wait()

	result := &Result{
		Stdout: output.String(),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
	}

	return result, nil
}

// StartBackground starts a command in the background
func (t *Tool) StartBackground(command string) (string, error) {
	if !t.config.AllowBackground {
		return "", fmt.Errorf("background processes not allowed")
	}

	id := fmt.Sprintf("bg-%d", time.Now().UnixNano())

	resolved := t.resolveRefs(command)
	cmd := exec.Command("sh", "-c", resolved)
	cmd.Dir = t.config.WorkDir
	cmd.Env = t.filterEnv(os.Environ())

	output := &bytes.Buffer{}
	cmd.Stdout = output
	cmd.Stderr = output

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start: %w", err)
	}

	bp := &BackgroundProcess{
		ID:        id,
		Command:   command,
		Cmd:       cmd,
		Output:    output,
		StartTime: time.Now(),
	}

	t.bgMu.Lock()
	t.bgProcesses[id] = bp
	t.bgMu.Unlock()

	// Monitor process completion
	go func() {
		err := cmd.Wait()
		t.bgMu.Lock()
		bp.Done = true
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				bp.ExitCode = exitErr.ExitCode()
			} else {
				bp.Error = err.Error()
			}
		}
		t.bgMu.Unlock()
	}()

	logger.Infow("started background process", "id", id, "command", command)
	return id, nil
}

// GetBackgroundStatus returns the status of a background process
func (t *Tool) GetBackgroundStatus(id string) (*BackgroundProcess, error) {
	t.bgMu.RLock()
	defer t.bgMu.RUnlock()

	bp, ok := t.bgProcesses[id]
	if !ok {
		return nil, fmt.Errorf("process not found: %s", id)
	}

	return bp, nil
}

// StopBackground stops a background process
func (t *Tool) StopBackground(id string) error {
	t.bgMu.Lock()
	defer t.bgMu.Unlock()

	bp, ok := t.bgProcesses[id]
	if !ok {
		return fmt.Errorf("process not found: %s", id)
	}

	if !bp.Done {
		if err := bp.Cmd.Process.Signal(syscall.SIGTERM); err != nil {
			bp.Cmd.Process.Kill()
		}
	}

	delete(t.bgProcesses, id)
	logger.Infow("stopped background process", "id", id)
	return nil
}

// ListBackground returns all background processes
func (t *Tool) ListBackground() []*BackgroundProcess {
	t.bgMu.RLock()
	defer t.bgMu.RUnlock()

	list := make([]*BackgroundProcess, 0, len(t.bgProcesses))
	for _, bp := range t.bgProcesses {
		list = append(list, bp)
	}
	return list
}

// filterEnv filters environment variables
func (t *Tool) filterEnv(env []string) []string {
	// If Whitelist is provided, exclusively use it
	if len(t.config.EnvWhitelist) > 0 {
		result := make([]string, 0)
		for _, e := range env {
			for _, allowed := range t.config.EnvWhitelist {
				if strings.HasPrefix(strings.ToUpper(e), strings.ToUpper(allowed)+"=") {
					result = append(result, e)
					break
				}
			}
		}
		return result
	}

	// Default blacklist behavior
	defaultFilter := []string{
		"AWS_SECRET", "ANTHROPIC_API_KEY", "OPENAI_API_KEY",
		"GOOGLE_API_KEY", "SLACK_BOT_TOKEN", "DISCORD_TOKEN",
		"TELEGRAM_BOT_TOKEN", "LANGO_PASSPHRASE",
	}
	filterList := append(defaultFilter, t.config.EnvFilter...)

	result := make([]string, 0, len(env))
	for _, e := range env {
		exclude := false
		for _, f := range filterList {
			if strings.HasPrefix(strings.ToUpper(e), strings.ToUpper(f)+"=") {
				exclude = true
				break
			}
		}
		if !exclude {
			result = append(result, e)
		}
	}
	return result
}

// Cleanup terminates all background processes
func (t *Tool) Cleanup() {
	t.bgMu.Lock()
	defer t.bgMu.Unlock()

	for id, bp := range t.bgProcesses {
		if !bp.Done {
			bp.Cmd.Process.Kill()
		}
		delete(t.bgProcesses, id)
	}
}
