// Package sandbox provides the `lango sandbox` CLI command group
// for inspecting OS-level process sandbox status and running smoke tests.
package sandbox

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/clihttp"
	"github.com/langoai/lango/internal/config"
	sandboxos "github.com/langoai/lango/internal/sandbox/os"
	"github.com/langoai/lango/internal/storage"
)

// BootLoader is the optional dependency that opens the encrypted application
// database so `lango sandbox status` can query the recent SandboxDecision
// audit trail. It is optional: if nil or if it returns an error (database
// locked, signed-out, missing), status renders without the Recent Decisions
// section so the command remains usable as a pure sandbox-layer diagnostic.
type BootLoader func() (*bootstrap.Result, error)

// NewSandboxCmd creates the top-level `lango sandbox` command.
//
// cfgLoader runs lightweight bootstrap and returns only the config (closing
// the DB on the way out) — used by `sandbox test` which does not need the
// audit DB, and as the graceful-degradation fallback for `sandbox status`
// when `bootLoader` is unavailable. bootLoader runs the full bootstrap and
// returns the result with initialized bootstrap storage — used by `sandbox status` to
// query the sandbox decision audit trail.
//
// `sandbox status` prefers bootLoader so that one bootstrap pass serves both
// the config rendering and the audit query (no double passphrase prompt).
// When bootLoader is nil or returns an error (signed-out, locked DB,
// non-interactive environments), status falls back to cfgLoader so the
// config / capabilities / backend availability sections still render —
// only the Recent Sandbox Decisions section is silently skipped.
func NewSandboxCmd(cfgLoader func() (*config.Config, error), bootLoader BootLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sandbox",
		Short: "Manage OS-level process sandbox",
		Long:  "Inspect sandbox configuration, platform capabilities, and run isolation smoke tests.",
	}

	cmd.AddCommand(newStatusCmd(cfgLoader, bootLoader))
	cmd.AddCommand(newTestCmd(cfgLoader))
	cmd.AddCommand(newProbeNetCmd())

	return cmd
}

// versioner is an optional interface implemented by isolators that can report
// a version string (e.g. BwrapIsolator captures `bwrap --version`). The test
// command type-asserts against this interface so backends without a meaningful
// version (Seatbelt, noop) simply omit the line.
type versioner interface {
	Version() string
}

// networkIsolator is an optional interface implemented by isolators that
// separately track network isolation capability (e.g. BwrapIsolator's
// two-phase smoke probe, where the base namespace probe can succeed while
// the --unshare-net probe fails). Status rendering uses this to surface
// partial degradation that would otherwise be invisible — users seeing
// "MCP works but exec/skill is rejected" would have no way to diagnose it.
type networkIsolator interface {
	NetworkIsolationAvailable() bool
	NetworkIsolationReason() string
}

// newStatusCmd creates `lango sandbox status`.
func newStatusCmd(cfgLoader func() (*config.Config, error), bootLoader BootLoader) *cobra.Command {
	var sessionPrefix string
	var output string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show sandbox configuration and platform capabilities",
		RunE: func(cmd *cobra.Command, _ []string) error {
			w := cmd.OutOrStdout()
			if err := validateStatusOutput(output); err != nil {
				return err
			}

			// Try bootLoader first so a single bootstrap pass serves both
			// the config rendering AND the Recent Decisions audit query
			// (no double passphrase prompt — that was the v1 regression).
			// On nil bootLoader or any failure, fall back to cfgLoader so
			// the config / capabilities / backend sections still render
			// in degraded modes (signed-out, locked DB, non-interactive).
			// Only the Recent Decisions section is silently skipped.
			var (
				cfg  *config.Config
				boot *bootstrap.Result
			)
			if bootLoader != nil {
				if b, err := bootLoader(); err == nil && b != nil && b.Config != nil {
					boot = b
					cfg = b.Config
					defer func() { _ = boot.Close() }()
				}
			}
			if cfg == nil {
				if cfgLoader == nil {
					return fmt.Errorf("sandbox status requires a config loader or bootstrap loader")
				}
				c, err := cfgLoader()
				if err != nil {
					return fmt.Errorf("load config: %w", err)
				}
				cfg = c
			}

			snapshot := buildStatusSnapshot(cmd.Context(), cfg, boot, sessionPrefix)
			return renderStatusSnapshot(w, snapshot, output)
		},
	}
	cmd.Flags().StringVar(&sessionPrefix, "session", "",
		"Filter Recent Sandbox Decisions by session key prefix (default: show global)")
	cmd.Flags().StringVar(&output, "output", "table", "Output format: table, json, or plain")
	return cmd
}

type statusSnapshot struct {
	SessionPrefix        string                     `json:"-"`
	Configuration        statusConfiguration        `json:"configuration"`
	ActiveIsolation      statusActiveIsolation      `json:"active_isolation"`
	PlatformCapabilities statusPlatformCapabilities `json:"platform_capabilities"`
	BackendAvailability  []statusBackend            `json:"backend_availability"`
	RecentDecisions      []statusDecision           `json:"recent_decisions"`
	Warnings             statusWarnings             `json:"warnings"`
}

type statusConfiguration struct {
	Enabled        bool   `json:"enabled"`
	FailClosed     bool   `json:"fail_closed"`
	FailMode       string `json:"fail_mode,omitempty"`
	Backend        string `json:"backend"`
	BackendLabel   string `json:"backend_label"`
	NetworkMode    string `json:"network_mode"`
	Workspace      string `json:"workspace"`
	ExplicitOptOut bool   `json:"explicit_opt_out"`
}

type statusActiveIsolation struct {
	Isolator                    string `json:"isolator"`
	Available                   bool   `json:"available"`
	Reason                      string `json:"reason,omitempty"`
	NetworkIsolationUnavailable bool   `json:"network_isolation_unavailable,omitempty"`
	NetworkIsolationReason      string `json:"network_isolation_reason,omitempty"`
}

type statusPlatformCapabilities struct {
	Platform    string `json:"platform"`
	Kernel      string `json:"kernel"`
	Seatbelt    string `json:"seatbelt"`
	Landlock    string `json:"landlock"`
	Seccomp     string `json:"seccomp"`
	HasSeatbelt bool   `json:"has_seatbelt"`
	HasLandlock bool   `json:"has_landlock"`
	LandlockABI int    `json:"landlock_abi"`
	HasSeccomp  bool   `json:"has_seccomp"`
}

type statusBackend struct {
	Name      string `json:"name"`
	Mode      string `json:"mode"`
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}

type statusDecision struct {
	Timestamp        string `json:"timestamp"`
	SessionKeyPrefix string `json:"session_key_prefix"`
	Decision         string `json:"decision"`
	Backend          string `json:"backend"`
	Target           string `json:"target"`
	Reason           string `json:"reason,omitempty"`
}

type statusWarnings struct {
	AllowedNetworkIPsMacOSOnly bool `json:"allowed_network_ips_macos_only"`
}

func validateStatusOutput(output string) error {
	switch strings.ToLower(strings.TrimSpace(output)) {
	case "table", "json", "plain":
		return nil
	default:
		return fmt.Errorf("unsupported output format %q (expected table, json, or plain)", output)
	}
}

func buildStatusSnapshot(ctx context.Context, cfg *config.Config, boot *bootstrap.Result, sessionPrefix string) statusSnapshot {
	workspacePath := cfg.Sandbox.WorkspacePath
	if workspacePath == "" {
		workspacePath, _ = os.Getwd()
	}

	mode, _ := sandboxos.ParseBackendMode(cfg.Sandbox.Backend)
	candidates := sandboxos.PlatformBackendCandidates()
	var isolator sandboxos.OSIsolator
	var info sandboxos.BackendInfo
	optedOut := cfg.Sandbox.Enabled && mode == sandboxos.BackendNone
	if cfg.Sandbox.Enabled && !optedOut {
		isolator, info = sandboxos.SelectBackend(mode, candidates)
	}
	status := sandboxos.NewSandboxStatus(cfg.Sandbox.Enabled, cfg.Sandbox.FailClosed, isolator)

	failMode := ""
	if status.Enabled && !optedOut {
		failMode = "fail-open (warning + unsandboxed execution)"
		if status.FailClosed {
			failMode = "fail-closed (execution rejected)"
		}
	}

	backendLabel := mode.String()
	if optedOut {
		backendLabel = "none (explicit opt-out - fail-closed not applied)"
	} else if mode == sandboxos.BackendAuto && info.Name != "" {
		backendLabel = fmt.Sprintf("auto (resolved: %s)", info.Name)
	}

	active := statusActiveIsolation{
		Isolator:  status.Isolator.Name(),
		Available: status.Isolator.Available(),
	}
	if !active.Available {
		active.Reason = status.Isolator.Reason()
	} else if ni, ok := status.Isolator.(networkIsolator); ok && !ni.NetworkIsolationAvailable() {
		active.NetworkIsolationUnavailable = true
		active.NetworkIsolationReason = ni.NetworkIsolationReason()
	}

	caps := status.Capabilities
	backends := make([]statusBackend, 0, len(candidates))
	for _, bi := range sandboxos.ListBackends(candidates) {
		backends = append(backends, statusBackend{
			Name:      bi.Name,
			Mode:      bi.Mode.String(),
			Available: bi.Available,
			Reason:    bi.Reason,
		})
	}

	return statusSnapshot{
		SessionPrefix: sessionPrefix,
		Configuration: statusConfiguration{
			Enabled:        status.Enabled,
			FailClosed:     status.FailClosed,
			FailMode:       failMode,
			Backend:        mode.String(),
			BackendLabel:   backendLabel,
			NetworkMode:    cfg.Sandbox.NetworkMode,
			Workspace:      workspacePath,
			ExplicitOptOut: optedOut,
		},
		ActiveIsolation: active,
		PlatformCapabilities: statusPlatformCapabilities{
			Platform:    caps.Platform,
			Kernel:      caps.KernelVersion,
			Seatbelt:    capabilityReasonStatus(caps.HasSeatbelt, caps.SeatbeltReason, caps.Platform, "darwin"),
			Landlock:    capabilityReasonStatus(caps.HasLandlock, caps.LandlockReason, caps.Platform, "linux"),
			Seccomp:     capabilityReasonStatus(caps.HasSeccomp, caps.SeccompReason, caps.Platform, "linux"),
			HasSeatbelt: caps.HasSeatbelt,
			HasLandlock: caps.HasLandlock,
			LandlockABI: caps.LandlockABI,
			HasSeccomp:  caps.HasSeccomp,
		},
		BackendAvailability: backends,
		RecentDecisions:     collectRecentDecisions(ctx, boot, sessionPrefix),
		Warnings: statusWarnings{
			AllowedNetworkIPsMacOSOnly: runtime.GOOS == "linux" && len(cfg.Sandbox.AllowedNetworkIPs) > 0,
		},
	}
}

func renderStatusSnapshot(w io.Writer, snapshot statusSnapshot, output string) error {
	switch strings.ToLower(strings.TrimSpace(output)) {
	case "json":
		return clihttp.PrintJSON(w, snapshot)
	case "plain":
		renderStatusPlain(w, snapshot)
		return nil
	default:
		renderStatusTable(w, snapshot)
		return nil
	}
}

func renderStatusTable(w io.Writer, snapshot statusSnapshot) {
	cfg := snapshot.Configuration
	fmt.Fprintln(w, "Sandbox Configuration:")
	fmt.Fprintf(w, "  Enabled:        %v\n", cfg.Enabled)
	if cfg.Enabled {
		if cfg.ExplicitOptOut {
			fmt.Fprintf(w, "  Backend:        none (explicit opt-out — fail-closed not applied)\n")
		} else {
			fmt.Fprintf(w, "  Fail-Closed:    %s\n", cfg.FailMode)
			fmt.Fprintf(w, "  Backend:        %s\n", cfg.BackendLabel)
		}
	}
	fmt.Fprintf(w, "  Network Mode:   %s\n", cfg.NetworkMode)
	fmt.Fprintf(w, "  Workspace:      %s\n", cfg.Workspace)

	active := snapshot.ActiveIsolation
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Active Isolation:")
	fmt.Fprintf(w, "  Isolator:       %s\n", active.Isolator)
	if !active.Available {
		fmt.Fprintf(w, "  Available:      false\n")
		fmt.Fprintf(w, "  Reason:         %s\n", active.Reason)
	} else {
		fmt.Fprintf(w, "  Available:      true\n")
		if active.NetworkIsolationUnavailable {
			fmt.Fprintf(w, "  Network Iso:    unavailable (%s)\n", active.NetworkIsolationReason)
		}
	}

	caps := snapshot.PlatformCapabilities
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Platform Capabilities:")
	fmt.Fprintf(w, "  Platform:       %s\n", caps.Platform)
	fmt.Fprintf(w, "  Kernel:         %s\n", caps.Kernel)
	fmt.Fprintf(w, "  Seatbelt:       %s\n", caps.Seatbelt)
	fmt.Fprintf(w, "  Landlock:       %s\n", caps.Landlock)
	fmt.Fprintf(w, "  seccomp:        %s\n", caps.Seccomp)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Backend Availability:")
	for _, bi := range snapshot.BackendAvailability {
		state := "available"
		if !bi.Available {
			state = fmt.Sprintf("unavailable (%s)", bi.Reason)
		}
		fmt.Fprintf(w, "  %-14s  %s\n", bi.Name+":", state)
	}

	if snapshot.Warnings.AllowedNetworkIPsMacOSOnly {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "WARNING: allowedNetworkIPs is macOS-only; Linux isolation is not yet enforced")
	}

	renderDecisionLines(w, snapshot.RecentDecisions, snapshot.SessionPrefix)
}

func renderStatusPlain(w io.Writer, snapshot statusSnapshot) {
	cfg := snapshot.Configuration
	fmt.Fprintf(w, "enabled=%v\n", cfg.Enabled)
	fmt.Fprintf(w, "fail_closed=%v\n", cfg.FailClosed)
	if cfg.FailMode != "" {
		fmt.Fprintf(w, "fail_mode=%s\n", cfg.FailMode)
	}
	fmt.Fprintf(w, "backend=%s\n", cfg.BackendLabel)
	fmt.Fprintf(w, "network_mode=%s\n", cfg.NetworkMode)
	fmt.Fprintf(w, "workspace=%s\n", cfg.Workspace)
	fmt.Fprintf(w, "isolator=%s\n", snapshot.ActiveIsolation.Isolator)
	fmt.Fprintf(w, "available=%v\n", snapshot.ActiveIsolation.Available)
	if snapshot.ActiveIsolation.Reason != "" {
		fmt.Fprintf(w, "reason=%s\n", snapshot.ActiveIsolation.Reason)
	}
	fmt.Fprintf(w, "platform=%s\n", snapshot.PlatformCapabilities.Platform)
	fmt.Fprintf(w, "kernel=%s\n", snapshot.PlatformCapabilities.Kernel)
	for _, backend := range snapshot.BackendAvailability {
		state := "available"
		if !backend.Available {
			state = "unavailable"
		}
		fmt.Fprintf(w, "backend_availability=%s:%s", backend.Name, state)
		if backend.Reason != "" {
			fmt.Fprintf(w, ":%s", backend.Reason)
		}
		fmt.Fprintln(w)
	}
	if snapshot.Warnings.AllowedNetworkIPsMacOSOnly {
		fmt.Fprintln(w, "warning=allowedNetworkIPs is macOS-only; Linux isolation is not yet enforced")
	}
}

// renderRecentDecisions queries the audit log for the most recent
// SandboxDecisionEvent records and prints them to w. It is best-effort:
// any failure (missing storage, schema unavailable, query error) is
// silently ignored so the diagnostic remains usable as a sandbox-layer
// inspection tool that does not depend on audit availability.
//
// The caller owns the *bootstrap.Result lifecycle; this helper does not close it.
func renderRecentDecisions(ctx context.Context, w io.Writer, boot *bootstrap.Result, sessionPrefix string) {
	renderDecisionLines(w, collectRecentDecisions(ctx, boot, sessionPrefix), sessionPrefix)
}

func collectRecentDecisions(ctx context.Context, boot *bootstrap.Result, sessionPrefix string) []statusDecision {
	if boot == nil || boot.Storage == nil {
		return []statusDecision{}
	}

	var records []storage.SandboxDecisionRecord
	decisions, err := boot.Storage.RecentSandboxDecisions(ctx, sessionPrefix, 10)
	if err == nil {
		records = decisions
	}
	if err != nil || len(records) == 0 {
		return []statusDecision{}
	}

	out := make([]statusDecision, 0, len(records))
	for _, d := range records {
		out = append(out, statusDecisionFromRecord(d))
	}
	return out
}

func statusDecisionFromRecord(d storage.SandboxDecisionRecord) statusDecision {
	backend := d.Backend
	if d.Decision != "applied" || backend == "" {
		backend = "-"
	}
	return statusDecision{
		Timestamp:        d.Timestamp.Format("2006-01-02 15:04:05"),
		SessionKeyPrefix: truncateSessionKey(d.SessionKey, 8),
		Decision:         d.Decision,
		Backend:          backend,
		Target:           d.Target,
		Reason:           d.Reason,
	}
}

func renderDecisionLines(w io.Writer, decisions []statusDecision, sessionPrefix string) {
	if len(decisions) == 0 {
		return
	}

	title := "Recent Sandbox Decisions (global, last 10):"
	if sessionPrefix != "" {
		title = fmt.Sprintf("Recent Sandbox Decisions (session=%s, last 10):", sessionPrefix)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, title)
	for _, d := range decisions {
		fmt.Fprintln(w, formatDecisionLineFromStatus(d))
	}
}

func formatDecisionLineFromStatus(d statusDecision) string {
	line := fmt.Sprintf("  %s  [%s] %-9s %-9s %s",
		d.Timestamp, d.SessionKeyPrefix, d.Decision, d.Backend, d.Target)
	if d.Reason != "" {
		line += fmt.Sprintf(" (%s)", d.Reason)
	}
	return line
}

// truncateSessionKey shortens long session keys for display, padding empty
// keys to a fixed width so columns align.
func truncateSessionKey(key string, width int) string {
	if key == "" {
		return strings.Repeat("-", width)
	}
	if len(key) <= width {
		return key + strings.Repeat(" ", width-len(key))
	}
	return key[:width]
}

// newTestCmd creates `lango sandbox test`.
func newTestCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	return newTestCmdWithBackendResolver(cfgLoader, sandboxos.PlatformBackendCandidates, sandboxos.SelectBackend)
}

func newTestCmdWithBackendResolver(
	cfgLoader func() (*config.Config, error),
	candidatesFunc func() []sandboxos.BackendCandidate,
	selectBackend func(sandboxos.BackendMode, []sandboxos.BackendCandidate) (sandboxos.OSIsolator, sandboxos.BackendInfo),
) *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Run OS sandbox smoke tests",
		Long: "Verify that the OS-level sandbox can restrict filesystem writes, allow reads, " +
			"permit workspace writes, and deny network connections.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			w := cmd.OutOrStdout()

			cfg, err := cfgLoader()
			if err != nil {
				return err
			}

			mode, _ := sandboxos.ParseBackendMode(cfg.Sandbox.Backend)
			if mode == sandboxos.BackendNone {
				fmt.Fprintln(w, "Sandbox backend explicitly set to none — no isolation to test")
				return nil
			}
			isolator, info := selectBackend(mode, candidatesFunc())
			if !isolator.Available() {
				fmt.Fprintf(w, "Sandbox backend %s not available: %s\n", info.Mode.String(), isolator.Reason())
				return nil
			}

			fmt.Fprintf(w, "Using isolator: %s (backend: %s)\n", isolator.Name(), info.Mode.String())
			if v, ok := isolator.(versioner); ok && v.Version() != "" {
				fmt.Fprintf(w, "Version: %s\n", v.Version())
			}
			fmt.Fprintln(w)

			tests := []struct {
				label  string
				passOK string
				failOK string
				run    func(sandboxos.OSIsolator) bool
			}{
				{
					label:  "Write restriction (deny /etc)",
					passOK: "PASS (write correctly denied)",
					failOK: "FAIL (write was not denied)",
					run:    runWriteTest,
				},
				{
					label:  "Read permission (allow system file)",
					passOK: "PASS (read succeeded)",
					failOK: "FAIL (read was denied)",
					run:    runReadTest,
				},
				{
					label:  "Workspace write (allow tmp dir)",
					passOK: "PASS (workspace write succeeded)",
					failOK: "FAIL (workspace write blocked)",
					run:    runWorkspaceWriteTest,
				},
				{
					label:  "Network deny (loopback unreachable)",
					passOK: "PASS (connect correctly denied)",
					failOK: "FAIL (sandboxed child reached host listener)",
					run:    runNetworkDenyTest,
				},
			}

			allOK := true
			for _, tt := range tests {
				fmt.Fprintf(w, "%-40s ... ", tt.label)
				ok := tt.run(isolator)
				if ok {
					fmt.Fprintln(w, tt.passOK)
				} else {
					fmt.Fprintln(w, tt.failOK)
					allOK = false
				}
			}

			fmt.Fprintln(w)
			if allOK {
				fmt.Fprintln(w, "All tests passed.")
			} else {
				fmt.Fprintln(w, "Some tests failed.")
			}

			return nil
		},
	}
}

// newProbeNetCmd creates the hidden `lango sandbox _probe-net <addr>` helper.
//
// It is used by runNetworkDenyTest to perform a sandboxed TCP connect attempt
// without depending on external tools (nc/curl/bash). The parent test opens an
// ephemeral loopback listener and re-invokes the lango binary as a sandboxed
// child to dial that address; if the sandbox blocks the connection (Seatbelt
// (deny network*) on macOS, --unshare-net on Linux bwrap) the child exits
// non-zero, which the parent reads as PASS.
func newProbeNetCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "_probe-net <addr>",
		Hidden: true,
		Short:  "internal: attempt a TCP connection (used by sandbox test)",
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			conn, err := net.DialTimeout("tcp", args[0], 2*time.Second)
			if err != nil {
				return err
			}
			_ = conn.Close()
			return nil
		},
	}
}

// readOnlyPolicy returns a sandbox policy that allows reading the entire
// filesystem but blocks all writes and network access.
func readOnlyPolicy() sandboxos.Policy {
	return sandboxos.Policy{
		Filesystem: sandboxos.FilesystemPolicy{
			ReadOnlyGlobal: true,
			WritePaths:     []string{"/tmp"},
		},
		Network: sandboxos.NetworkDeny,
		Process: sandboxos.ProcessPolicy{AllowFork: true},
	}
}

// discardOutput silences a command's stdout and stderr without using shell
// redirection to /dev/null. The parent-side io.Discard avoids opening
// /dev/null inside the sandbox (which Seatbelt's default-deny would block,
// causing false negatives in the smoke tests). The child inherits the pipe
// FDs that exec.Cmd creates for non-*os.File writers, and those FDs are
// already open before the sandbox takes effect.
func discardOutput(c *exec.Cmd) {
	c.Stdout = io.Discard
	c.Stderr = io.Discard
}

// findTouch locates the touch binary across common Linux/macOS layouts:
// PATH lookup first, then explicit fallbacks for non-merged-/usr images
// (BusyBox/Alpine ship touch under /bin, not /usr/bin). Returns the
// absolute path or an empty string when no touch is available — callers
// must treat empty as "test inconclusive" rather than as a sandbox failure
// (otherwise an ENOENT from a missing binary can produce a false PASS).
func findTouch() string {
	if p, err := exec.LookPath("touch"); err == nil {
		return p
	}
	for _, candidate := range []string{"/usr/bin/touch", "/bin/touch"} {
		if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
			return candidate
		}
	}
	return ""
}

// runWriteTest attempts to write to a restricted path under sandbox and
// returns true if the write was correctly blocked.
func runWriteTest(isolator sandboxos.OSIsolator) bool {
	touch := findTouch()
	if touch == "" {
		// Without touch we cannot distinguish "sandbox blocked the write"
		// from "binary missing from PATH". Refuse to declare PASS in that
		// case so the smoke test does not produce a false positive.
		return false
	}
	c := exec.Command(touch, "/etc/lango-sandbox-test")
	discardOutput(c)
	if err := isolator.Apply(context.Background(), c, readOnlyPolicy()); err != nil {
		return false
	}
	// The command should fail (permission denied).
	return c.Run() != nil
}

// runReadTest attempts to read a file under sandbox and
// returns true if the read succeeded.
func runReadTest(isolator sandboxos.OSIsolator) bool {
	target := readTestPath()
	c := exec.Command("/bin/cat", target)
	discardOutput(c)
	if err := isolator.Apply(context.Background(), c, readOnlyPolicy()); err != nil {
		return false
	}
	return c.Run() == nil
}

// runWorkspaceWriteTest attempts to write a file inside a temporary workspace
// directory that the sandbox policy explicitly allows. Returns true when the
// write succeeds (allowed paths must remain writable).
//
// macOS quirk: os.MkdirTemp returns paths under /var/folders/... but the real
// path is /private/var/folders/... and Seatbelt resolves subpaths against the
// real path. We resolve via filepath.EvalSymlinks before passing to the policy.
func runWorkspaceWriteTest(isolator sandboxos.OSIsolator) bool {
	work, err := os.MkdirTemp("", "lango-sandbox-ws-*")
	if err != nil {
		return false
	}
	defer os.RemoveAll(work)

	resolved, err := filepath.EvalSymlinks(work)
	if err != nil {
		resolved = work
	}

	touch := findTouch()
	if touch == "" {
		// Cannot exercise the workspace-write path without a touch binary.
		// Treat as inconclusive; the test runner reports FAIL but the
		// reason is environmental, not a sandbox regression.
		return false
	}

	target := filepath.Join(resolved, "probe.txt")
	c := exec.Command(touch, target)
	discardOutput(c)
	policy := sandboxos.Policy{
		Filesystem: sandboxos.FilesystemPolicy{
			ReadOnlyGlobal: true,
			WritePaths:     []string{resolved, "/tmp"},
		},
		Network: sandboxos.NetworkDeny,
		Process: sandboxos.ProcessPolicy{AllowFork: true},
	}
	if err := isolator.Apply(context.Background(), c, policy); err != nil {
		return false
	}
	if err := c.Run(); err != nil {
		return false
	}
	_, err = os.Stat(target)
	return err == nil
}

// runNetworkDenyTest verifies that the sandbox blocks outbound TCP connects
// even to a known-reachable loopback endpoint. The test:
//  1. opens an ephemeral TCP listener on 127.0.0.1 in the parent process
//     (so we have a target the host could otherwise reach with certainty);
//  2. re-invokes the lango binary as a sandboxed child via the hidden
//     `_probe-net <addr>` subcommand, which calls net.DialTimeout;
//  3. returns true (PASS) only if the child failed to connect.
//
// External tools (nc/curl/bash) are intentionally not used so the test runs
// in minimal Docker images. The deterministic loopback target ensures any
// failure is attributable to the sandbox, not to host network conditions.
func runNetworkDenyTest(isolator sandboxos.OSIsolator) bool {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return false
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()

	target := ln.Addr().String()
	self, err := os.Executable()
	if err != nil {
		return false
	}

	c := exec.Command(self, "sandbox", "_probe-net", target)
	discardOutput(c)
	if err := isolator.Apply(context.Background(), c, readOnlyPolicy()); err != nil {
		return false
	}
	// PASS if the child failed to connect (sandbox blocked it).
	return c.Run() != nil
}

// readTestPath returns a readable file path suitable for the current platform.
func readTestPath() string {
	if runtime.GOOS == "darwin" {
		return "/etc/hosts"
	}
	return "/etc/hostname"
}

// capabilityReasonStatus formats a capability's status with a reason string.
// Reasons containing "not yet implemented" are shown as "unknown" (defensive
// against future stub probes); definitive results (e.g., "Landlock ABI 3",
// "Landlock not supported by this kernel") are shown as "available" or
// "unavailable" with the qualified reason inline.
func capabilityReasonStatus(available bool, reason, currentPlatform, requiredPlatform string) string {
	if available {
		if reason != "" {
			return fmt.Sprintf("available (%s)", reason)
		}
		return "available"
	}
	if !strings.EqualFold(currentPlatform, requiredPlatform) {
		return fmt.Sprintf("n/a (not on %s)", requiredPlatform)
	}
	if strings.Contains(reason, "not yet implemented") {
		return fmt.Sprintf("unknown (%s)", reason)
	}
	if reason != "" {
		return fmt.Sprintf("unavailable (%s)", reason)
	}
	return "unavailable"
}
