// Package sandbox provides the `lango sandbox` CLI command group
// for inspecting OS-level process sandbox status and running smoke tests.
package sandbox

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
	sandboxos "github.com/langoai/lango/internal/sandbox/os"
)

// NewSandboxCmd creates the top-level `lango sandbox` command.
func NewSandboxCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sandbox",
		Short: "Manage OS-level process sandbox",
		Long:  "Inspect sandbox configuration, platform capabilities, and run isolation smoke tests.",
	}

	cmd.AddCommand(newStatusCmd(cfgLoader))
	cmd.AddCommand(newTestCmd(cfgLoader))

	return cmd
}

// newStatusCmd creates `lango sandbox status`.
func newStatusCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	return &cobra.Command{
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

			return nil
		},
	}
}

// newTestCmd creates `lango sandbox test`.
func newTestCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Run OS sandbox smoke tests",
		Long:  "Verify that the OS-level sandbox can restrict filesystem writes and allow reads.",
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

			fmt.Fprintf(w, "Using isolator: %s (backend: %s)\n\n", isolator.Name(), info.Mode.String())

			// Test 1: verify sandbox blocks writes to a restricted path.
			fmt.Fprint(w, "Write restriction test ... ")
			writeOK := runWriteTest(isolator)
			if writeOK {
				fmt.Fprintln(w, "PASS (write correctly denied)")
			} else {
				fmt.Fprintln(w, "FAIL (write was not denied)")
			}

			// Test 2: verify sandbox allows reading system files.
			fmt.Fprint(w, "Read permission test   ... ")
			readOK := runReadTest(isolator)
			if readOK {
				fmt.Fprintln(w, "PASS (read succeeded)")
			} else {
				fmt.Fprintln(w, "FAIL (read was denied)")
			}

			fmt.Fprintln(w)
			if writeOK && readOK {
				fmt.Fprintln(w, "All tests passed.")
			} else {
				fmt.Fprintln(w, "Some tests failed.")
			}

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

// runWriteTest attempts to write to a restricted path under sandbox and
// returns true if the write was correctly blocked.
func runWriteTest(isolator sandboxos.OSIsolator) bool {
	c := exec.Command("/bin/sh", "-c", "touch /etc/lango-sandbox-test 2>/dev/null")
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
	c := exec.Command("/bin/sh", "-c", "cat "+target+" >/dev/null 2>&1")
	if err := isolator.Apply(context.Background(), c, readOnlyPolicy()); err != nil {
		return false
	}
	return c.Run() == nil
}

// readTestPath returns a readable file path suitable for the current platform.
func readTestPath() string {
	if runtime.GOOS == "darwin" {
		return "/etc/hosts"
	}
	return "/etc/hostname"
}

// capabilityReasonStatus formats a capability's status with a reason string.
// "probe not yet implemented" reasons are shown as "unknown"; definitive negative
// results (e.g., "sandbox-exec not found in PATH") are shown as "unavailable".
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
