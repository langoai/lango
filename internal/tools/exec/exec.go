package exec

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/logging"
	sandboxos "github.com/langoai/lango/internal/sandbox/os"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/session"
)

var logger = logging.SubsystemSugar("tool.exec")

// Config holds exec tool configuration
type Config struct {
	DefaultTimeout   time.Duration
	AllowBackground  bool
	WorkDir          string
	EnvFilter        []string             // environment variables to exclude
	EnvWhitelist     []string             // if set, ONLY these vars are allowed
	Refs             *security.RefStore   // secret reference token resolver
	OSIsolator       sandboxos.OSIsolator // OS-level sandbox (nil = disabled)
	SandboxPolicy    sandboxos.Policy     // policy for sandboxed execution
	FailClosed       bool                 // if true, reject execution when sandbox unavailable
	ExcludedCommands []string             // command basenames that bypass the sandbox
	Bus              *eventbus.Bus        // event bus for SandboxDecisionEvent (optional)
}

// Tool provides shell command execution
type Tool struct {
	config       Config
	bgProcesses  map[string]*BackgroundProcess
	bgMu         sync.RWMutex
	fallbackOnce sync.Once // emit the fail-open warning to stderr at most once per process
}

// syncBuffer is a thread-safe wrapper around bytes.Buffer.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (sb *syncBuffer) Write(p []byte) (int, error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.Write(p)
}

func (sb *syncBuffer) String() string {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.String()
}

// BackgroundProcess represents a running background command
type BackgroundProcess struct {
	ID        string
	Command   string
	Cmd       *exec.Cmd
	Output    *syncBuffer
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

// SetEventBus attaches an event bus for SandboxDecisionEvent publishing.
// Wiring may call this after Tool construction once the bus is available.
// Passing nil disables publishing (PublishSandboxDecision is a no-op on nil).
func (t *Tool) SetEventBus(bus *eventbus.Bus) {
	t.config.Bus = bus
}

// applySandbox applies OS-level sandbox to the command if configured.
// Returns a non-nil error only when fail-closed is set and the sandbox
// cannot be applied. In fail-open mode it logs a warning, emits a one-time
// stderr message, and returns nil.
//
// userCommand is the raw user command string (before sh -c wrapping and
// before secret token resolution). It is used both for ExcludedCommands
// matching and as the audit Command field.
func (t *Tool) applySandbox(ctx context.Context, cmd *exec.Cmd, userCommand string) error {
	if matched, pattern := excludedMatch(userCommand, t.config.ExcludedCommands); pattern != "" {
		t.publishDecision(ctx, userCommand, "excluded", "", pattern)
		logger.Warnw("sandbox bypassed: excluded command",
			"command", matched, "pattern", pattern)
		return nil
	}

	if t.config.OSIsolator == nil {
		if t.config.FailClosed {
			t.publishDecision(ctx, userCommand, "rejected", "no isolator configured", "")
			return fmt.Errorf("%w: no OS isolator configured", sandboxos.ErrSandboxRequired)
		}
		t.publishDecision(ctx, userCommand, "skipped", "no isolator configured", "")
		t.warnFallbackOnce("no isolator configured")
		return nil
	}
	if err := t.config.OSIsolator.Apply(ctx, cmd, t.config.SandboxPolicy); err != nil {
		if t.config.FailClosed {
			t.publishDecision(ctx, userCommand, "rejected", err.Error(), "")
			return fmt.Errorf("%w: %w", sandboxos.ErrSandboxRequired, err)
		}
		logger.Warnw("OS sandbox unavailable, proceeding without isolation", "error", err)
		t.publishDecision(ctx, userCommand, "skipped", err.Error(), "")
		t.warnFallbackOnce(err.Error())
		return nil
	}
	t.publishDecision(ctx, userCommand, "applied", "", "")
	return nil
}

// publishDecision builds and publishes a SandboxDecisionEvent. SessionKey is
// derived from ctx so that re-entry under different sessions produces correct
// audit attribution. The bus may be nil (publish is a no-op).
func (t *Tool) publishDecision(ctx context.Context, userCommand, decision, reason, pattern string) {
	backend := ""
	if t.config.OSIsolator != nil {
		backend = t.config.OSIsolator.Name()
	}
	eventbus.PublishSandboxDecision(t.config.Bus, eventbus.SandboxDecisionEvent{
		SessionKey: session.SessionKeyFromContext(ctx),
		Source:     "exec",
		Command:    userCommand,
		Decision:   decision,
		Backend:    backend,
		Reason:     reason,
		Pattern:    pattern,
	})
}

// warnFallbackOnce prints a one-shot stderr warning so the user notices that
// the sandbox is not active for this process. Subsequent calls are no-ops to
// keep agent output clean during long-running sessions.
func (t *Tool) warnFallbackOnce(reason string) {
	t.fallbackOnce.Do(func() {
		fmt.Fprintf(os.Stderr,
			"lango: WARNING — sandbox fallback active (reason: %s); commands run unsandboxed\n",
			reason)
	})
}

// excludedMatch returns the matched basename and pattern, or "", "" when no
// match. Matches against the basename of the first whitespace-separated token
// of the user command. cmd.Args[0] is "sh" because exec.Tool wraps commands
// in `sh -c`, so the user command must be parsed before sh wrapping.
//
// Conservative semantics: shell chains like "cd /tmp && git status" yield
// first token "cd" — they do NOT match excluded=["git"]. Bypass works only
// for direct invocations such as "git status" or "/usr/bin/git push".
func excludedMatch(userCommand string, patterns []string) (matched, pattern string) {
	if len(patterns) == 0 {
		return "", ""
	}
	fields := strings.Fields(userCommand)
	if len(fields) == 0 {
		return "", ""
	}
	base := filepath.Base(fields[0])
	for _, p := range patterns {
		if base == p {
			return base, p
		}
	}
	return "", ""
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

	if err := t.applySandbox(ctx, cmd, command); err != nil {
		return nil, err
	}

	logger.Infow("executing command", "command", command, "timeout", timeout)

	err := cmd.Run()
	sandboxos.CleanupProfileFile(cmd)

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

	if err := t.applySandbox(ctx, cmd, command); err != nil {
		return nil, err
	}

	// Start with PTY
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("start PTY: %w", err)
	}
	defer ptmx.Close()
	defer sandboxos.CleanupProfileFile(cmd)

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
		_ = cmd.Process.Signal(syscall.SIGTERM)
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

	output := &syncBuffer{}
	cmd.Stdout = output
	cmd.Stderr = output

	if err := t.applySandbox(context.Background(), cmd, command); err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("start background process: %w", err)
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
			_ = bp.Cmd.Process.Kill()
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
			_ = bp.Cmd.Process.Kill()
		}
		delete(t.bgProcesses, id)
	}
}
