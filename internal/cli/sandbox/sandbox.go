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
	cmd.AddCommand(newTestCmd())

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

			fmt.Fprintln(w, "Sandbox Configuration:")
			fmt.Fprintf(w, "  Enabled:        %v\n", cfg.Sandbox.Enabled)
			fmt.Fprintf(w, "  Fail-Closed:    %v\n", cfg.Sandbox.FailClosed)
			fmt.Fprintf(w, "  Network Mode:   %s\n", cfg.Sandbox.NetworkMode)
			fmt.Fprintf(w, "  Workspace:      %s\n", workspacePath)

			// Platform capabilities.
			caps := sandboxos.Probe()

			fmt.Fprintln(w)
			fmt.Fprintln(w, "Platform Capabilities:")
			fmt.Fprintf(w, "  Platform:       %s\n", caps.Platform)
			fmt.Fprintf(w, "  Kernel:         %s\n", caps.KernelVersion)
			fmt.Fprintf(w, "  Seatbelt:       %s\n", capabilityStatus(caps.HasSeatbelt, caps.Platform, "darwin"))
			fmt.Fprintf(w, "  Landlock:       %s\n", capabilityStatus(caps.HasLandlock, caps.Platform, "linux"))
			fmt.Fprintf(w, "  seccomp:        %s\n", capabilityStatus(caps.HasSeccomp, caps.Platform, "linux"))

			// Active isolator.
			isolator := sandboxos.NewOSIsolator()
			fmt.Fprintln(w)
			fmt.Fprintf(w, "Active Isolator:  %s\n", isolator.Name())

			// Platform-specific warnings.
			if runtime.GOOS == "linux" && len(cfg.Sandbox.AllowedNetworkIPs) > 0 {
				fmt.Fprintln(w)
				fmt.Fprintln(w, "WARNING: allowedNetworkIPs is macOS-only, ignored on Linux")
			}

			return nil
		},
	}
}

// newTestCmd creates `lango sandbox test`.
func newTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Run OS sandbox smoke tests",
		Long:  "Verify that the OS-level sandbox can restrict filesystem writes and allow reads.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			w := cmd.OutOrStdout()

			isolator := sandboxos.NewOSIsolator()
			if !isolator.Available() {
				fmt.Fprintln(w, "OS sandbox not available on this platform")
				return nil
			}

			fmt.Fprintf(w, "Using isolator: %s\n\n", isolator.Name())

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

// capabilityStatus formats a capability's availability for display.
func capabilityStatus(available bool, currentPlatform, requiredPlatform string) string {
	if available {
		return "available"
	}
	if !strings.EqualFold(currentPlatform, requiredPlatform) {
		return fmt.Sprintf("unavailable (%s)", currentPlatform)
	}
	return "unavailable"
}
