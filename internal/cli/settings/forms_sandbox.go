package settings

import (
	"strings"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
)

// NewOSSandboxForm creates the OS Sandbox configuration form.
// Field keys use the "os_sandbox_" prefix to avoid collisions with
// the P2P sandbox form which uses the "sandbox_" prefix.
func NewOSSandboxForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("OS Sandbox Configuration")

	enabled := &tuicore.Field{
		Key: "os_sandbox_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Sandbox.Enabled,
		Description: "Apply OS-level kernel sandbox (Seatbelt on macOS; Linux: planned, not yet enforced)",
	}
	form.AddField(enabled)
	isEnabled := func() bool { return enabled.Checked }

	form.AddField(&tuicore.Field{
		Key: "os_sandbox_fail_closed", Label: "  Fail-Closed Mode", Type: tuicore.InputBool,
		Checked:     cfg.Sandbox.FailClosed,
		Description: "Reject tool execution when OS sandbox is unavailable (default: fail-open)",
		VisibleWhen: isEnabled,
	})

	backend := cfg.Sandbox.Backend
	if backend == "" {
		backend = "auto"
	}
	form.AddField(&tuicore.Field{
		Key: "os_sandbox_backend", Label: "  Backend", Type: tuicore.InputSelect,
		Value:       backend,
		Options:     []string{"auto", "seatbelt", "bwrap", "native", "none"},
		Description: "Isolation backend: auto selects best available; bwrap/native planned, not yet implemented",
		VisibleWhen: isEnabled,
	})

	form.AddField(&tuicore.Field{
		Key: "os_sandbox_workspace_path", Label: "  Workspace Path", Type: tuicore.InputText,
		Value:       cfg.Sandbox.WorkspacePath,
		Placeholder: "(uses CWD when empty)",
		Description: "Root directory for workspace-relative write access",
		VisibleWhen: isEnabled,
	})

	networkMode := cfg.Sandbox.NetworkMode
	if networkMode == "" {
		networkMode = "deny"
	}
	form.AddField(&tuicore.Field{
		Key: "os_sandbox_network_mode", Label: "  Network Mode", Type: tuicore.InputSelect,
		Value:       networkMode,
		Options:     []string{"deny", "allow"},
		Description: "Network access for sandboxed processes (Linux: not yet enforced)",
		VisibleWhen: isEnabled,
	})

	form.AddField(&tuicore.Field{
		Key: "os_sandbox_allowed_ips", Label: "  Allowed Network IPs", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Sandbox.AllowedNetworkIPs, ","),
		Placeholder: "192.168.1.1,10.0.0.1 (comma-separated)",
		Description: "macOS only — IP addresses permitted for outbound connections",
		VisibleWhen: isEnabled,
	})

	form.AddField(&tuicore.Field{
		Key: "os_sandbox_allowed_write_paths", Label: "  Allowed Write Paths", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Sandbox.AllowedWritePaths, ","),
		Placeholder: "/tmp (comma-separated)",
		Description: "Additional paths writable from the sandbox (beyond workspace)",
		VisibleWhen: isEnabled,
	})

	form.AddField(&tuicore.Field{
		Key: "os_sandbox_timeout", Label: "  Timeout Per Tool", Type: tuicore.InputText,
		Value:       cfg.Sandbox.TimeoutPerTool.String(),
		Placeholder: "30s",
		Description: "Maximum execution time for a single sandboxed tool invocation",
		VisibleWhen: isEnabled,
	})

	seccompProfile := cfg.Sandbox.OS.SeccompProfile
	if seccompProfile == "" {
		seccompProfile = "moderate"
	}
	form.AddField(&tuicore.Field{
		Key: "os_sandbox_seccomp_profile", Label: "  seccomp Profile", Type: tuicore.InputSelect,
		Value:       seccompProfile,
		Options:     []string{"strict", "moderate", "permissive"},
		Description: "Linux only — not yet enforced",
		VisibleWhen: isEnabled,
	})

	form.AddField(&tuicore.Field{
		Key: "os_sandbox_seatbelt_profile", Label: "  Custom Seatbelt Profile", Type: tuicore.InputText,
		Value:       cfg.Sandbox.OS.SeatbeltCustomProfile,
		Placeholder: "(auto-generated when empty)",
		Description: "macOS only — path to a custom .sb profile file",
		VisibleWhen: isEnabled,
	})

	return &form
}
