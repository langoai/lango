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

	"entgo.io/ent/dialect/sql"
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent/auditlog"
	sandboxos "github.com/langoai/lango/internal/sandbox/os"
)

// BootLoader is the optional dependency that opens the encrypted application
// database so `lango sandbox status` can query the recent SandboxDecision
// audit trail. It is optional: if nil or if it returns an error (database
// locked, signed-out, missing), status renders without the Recent Decisions
// section so the command remains usable as a pure sandbox-layer diagnostic.
type BootLoader func() (*bootstrap.Result, error)

// NewSandboxCmd creates the top-level `lango sandbox` command.
// bootLoader is optional and may be nil — when present it enables the
// "Recent Sandbox Decisions" section in `sandbox status`.
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

// newStatusCmd creates `lango sandbox status`.
func newStatusCmd(cfgLoader func() (*config.Config, error), bootLoader BootLoader) *cobra.Command {
	var sessionPrefix string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show sandbox configuration and platform capabilities",
		RunE: func(cmd *cobra.Command, _ []string) error {
			w := cmd.OutOrStdout()

			cfg, err := cfgLoader()
			if err != nil {
				return err
			}

			// Sandbox configuration.
			workspacePath := cfg.Sandbox.WorkspacePath
			if workspacePath == "" {
				workspacePath, _ = os.Getwd()
			}

			// Resolve backend.
			mode, _ := sandboxos.ParseBackendMode(cfg.Sandbox.Backend)
			candidates := sandboxos.PlatformBackendCandidates()
			var isolator sandboxos.OSIsolator
			var info sandboxos.BackendInfo
			// backend=none is an explicit opt-out: runtime skips fail-closed,
			// so status reflects "no isolation" instead of building an isolator.
			optedOut := cfg.Sandbox.Enabled && mode == sandboxos.BackendNone
			if cfg.Sandbox.Enabled && !optedOut {
				isolator, info = sandboxos.SelectBackend(mode, candidates)
			}
			status := sandboxos.NewSandboxStatus(cfg.Sandbox.Enabled, cfg.Sandbox.FailClosed, isolator)

			fmt.Fprintln(w, "Sandbox Configuration:")
			fmt.Fprintf(w, "  Enabled:        %v\n", status.Enabled)
			if status.Enabled {
				if optedOut {
					fmt.Fprintf(w, "  Backend:        none (explicit opt-out — fail-closed not applied)\n")
				} else {
					failMode := "fail-open (warning + unsandboxed execution)"
					if status.FailClosed {
						failMode = "fail-closed (execution rejected)"
					}
					fmt.Fprintf(w, "  Fail-Closed:    %s\n", failMode)
					backendLabel := mode.String()
					if mode == sandboxos.BackendAuto && info.Name != "" {
						backendLabel = fmt.Sprintf("auto (resolved: %s)", info.Name)
					}
					fmt.Fprintf(w, "  Backend:        %s\n", backendLabel)
				}
			}
			fmt.Fprintf(w, "  Network Mode:   %s\n", cfg.Sandbox.NetworkMode)
			fmt.Fprintf(w, "  Workspace:      %s\n", workspacePath)

			// Active isolation.
			fmt.Fprintln(w)
			fmt.Fprintln(w, "Active Isolation:")
			fmt.Fprintf(w, "  Isolator:       %s\n", status.Isolator.Name())
			if !status.Isolator.Available() {
				fmt.Fprintf(w, "  Available:      false\n")
				fmt.Fprintf(w, "  Reason:         %s\n", status.Isolator.Reason())
			} else {
				fmt.Fprintf(w, "  Available:      true\n")
			}

			// Platform capabilities.
			caps := status.Capabilities
			fmt.Fprintln(w)
			fmt.Fprintln(w, "Platform Capabilities:")
			fmt.Fprintf(w, "  Platform:       %s\n", caps.Platform)
			fmt.Fprintf(w, "  Kernel:         %s\n", caps.KernelVersion)
			fmt.Fprintf(w, "  Seatbelt:       %s\n", capabilityReasonStatus(caps.HasSeatbelt, caps.SeatbeltReason, caps.Platform, "darwin"))
			fmt.Fprintf(w, "  Landlock:       %s\n", capabilityReasonStatus(caps.HasLandlock, caps.LandlockReason, caps.Platform, "linux"))
			fmt.Fprintf(w, "  seccomp:        %s\n", capabilityReasonStatus(caps.HasSeccomp, caps.SeccompReason, caps.Platform, "linux"))

			// Backend availability.
			fmt.Fprintln(w)
			fmt.Fprintln(w, "Backend Availability:")
			for _, bi := range sandboxos.ListBackends(candidates) {
				state := "available"
				if !bi.Available {
					state = fmt.Sprintf("unavailable (%s)", bi.Reason)
				}
				fmt.Fprintf(w, "  %-14s  %s\n", bi.Name+":", state)
			}

			// Platform-specific warnings.
			if runtime.GOOS == "linux" && len(cfg.Sandbox.AllowedNetworkIPs) > 0 {
				fmt.Fprintln(w)
				fmt.Fprintln(w, "WARNING: allowedNetworkIPs is macOS-only; Linux isolation is not yet enforced")
			}

			// Recent Sandbox Decisions (graceful — skip if audit DB unavailable).
			if bootLoader != nil {
				renderRecentDecisions(cmd.Context(), w, bootLoader, sessionPrefix)
			}

			return nil
		},
	}
	cmd.Flags().StringVar(&sessionPrefix, "session", "",
		"Filter Recent Sandbox Decisions by session key prefix (default: show global)")
	return cmd
}

// renderRecentDecisions queries the audit log for the most recent
// SandboxDecisionEvent records and prints them to w. It is best-effort:
// any failure (DB locked, signed-out, missing, schema unavailable) is
// silently ignored so the diagnostic remains usable as a sandbox-layer
// inspection tool that does not depend on audit availability.
func renderRecentDecisions(ctx context.Context, w io.Writer, bootLoader BootLoader, sessionPrefix string) {
	if bootLoader == nil {
		return
	}
	boot, err := bootLoader()
	if err != nil || boot == nil || boot.DBClient == nil {
		return
	}
	// Do NOT close the DB client here — it is owned by the bootstrap result
	// and the caller (cobra root) is responsible for the process lifecycle.
	// Closing here would break subsequent commands that share the same boot.

	q := boot.DBClient.AuditLog.Query().
		Where(auditlog.ActionEQ(auditlog.ActionSandboxDecision)).
		Order(auditlog.ByTimestamp(sql.OrderDesc())).
		Limit(10)
	if sessionPrefix != "" {
		q = q.Where(auditlog.SessionKeyHasPrefix(sessionPrefix))
	}
	decisions, err := q.All(ctx)
	if err != nil || len(decisions) == 0 {
		return
	}

	title := "Recent Sandbox Decisions (global, last 10):"
	if sessionPrefix != "" {
		title = fmt.Sprintf("Recent Sandbox Decisions (session=%s, last 10):", sessionPrefix)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, title)
	for _, d := range decisions {
		var decision, backend, reason string
		if v, ok := d.Details["decision"].(string); ok {
			decision = v
		}
		if v, ok := d.Details["backend"].(string); ok {
			backend = v
		}
		if backend == "" {
			backend = "-"
		}
		if v, ok := d.Details["reason"].(string); ok {
			reason = v
		}
		sessShort := truncateSessionKey(d.SessionKey, 8)
		fmt.Fprintf(w, "  %s  [%s] %-9s %-9s %s",
			d.Timestamp.Format("2006-01-02 15:04:05"),
			sessShort, decision, backend, d.Target)
		if reason != "" {
			fmt.Fprintf(w, " (%s)", reason)
		}
		fmt.Fprintln(w)
	}
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
			isolator, info := sandboxos.SelectBackend(mode, sandboxos.PlatformBackendCandidates())
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

// runWriteTest attempts to write to a restricted path under sandbox and
// returns true if the write was correctly blocked.
func runWriteTest(isolator sandboxos.OSIsolator) bool {
	c := exec.Command("/usr/bin/touch", "/etc/lango-sandbox-test")
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

	target := filepath.Join(resolved, "probe.txt")
	c := exec.Command("/usr/bin/touch", target)
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
