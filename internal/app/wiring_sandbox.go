package app

import (
	"github.com/langoai/lango/internal/config"
	sandboxos "github.com/langoai/lango/internal/sandbox/os"
)

// initOSSandbox creates an OS-level sandbox isolator from config using the backend registry.
// Returns nil if sandbox is disabled or the user explicitly opted out via backend=none.
// Callers MUST treat nil as "no sandbox wiring required" — fail-closed does not apply.
func initOSSandbox(cfg *config.Config) sandboxos.OSIsolator {
	if !cfg.Sandbox.Enabled {
		return nil
	}

	// Backend validity was checked by config.Validate; ParseBackendMode here is
	// infallible at runtime. Any error indicates a validation bypass bug.
	mode, _ := sandboxos.ParseBackendMode(cfg.Sandbox.Backend)

	// backend=none is an explicit opt-out. Treat it like !enabled so that
	// fail-closed does not reject execution.
	if mode == sandboxos.BackendNone {
		logger().Infow("OS sandbox disabled via backend=none (explicit opt-out)")
		return nil
	}

	candidates := sandboxos.PlatformBackendCandidates()
	iso, info := sandboxos.SelectBackend(mode, candidates)
	status := sandboxos.NewSandboxStatus(cfg.Sandbox.Enabled, cfg.Sandbox.FailClosed, iso)

	if !iso.Available() {
		if cfg.Sandbox.FailClosed {
			logger().Warnw("OS sandbox required but unavailable — tool execution will be blocked",
				"platform", status.Capabilities.Platform,
				"backend", info.Mode.String(),
				"reason", iso.Reason(),
				"capabilities", status.Capabilities.Summary())
		} else {
			logger().Warnw("OS sandbox enabled but unavailable — proceeding without isolation",
				"platform", status.Capabilities.Platform,
				"backend", info.Mode.String(),
				"reason", iso.Reason(),
				"capabilities", status.Capabilities.Summary())
		}
	} else {
		logger().Infow("OS sandbox initialized",
			"isolator", iso.Name(),
			"backend", info.Mode.String(),
			"networkMode", cfg.Sandbox.NetworkMode,
			"failClosed", cfg.Sandbox.FailClosed)
	}

	return iso
}

