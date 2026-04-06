package app

import (
	"github.com/langoai/lango/internal/config"
	sandboxos "github.com/langoai/lango/internal/sandbox/os"
)

// initOSSandbox creates an OS-level sandbox isolator from config.
// Returns nil if sandbox is disabled.
func initOSSandbox(cfg *config.Config) sandboxos.OSIsolator {
	if !cfg.Sandbox.Enabled {
		return nil
	}

	iso := sandboxos.NewOSIsolator()
	status := sandboxos.NewSandboxStatus(cfg.Sandbox.Enabled, cfg.Sandbox.FailClosed, iso)

	if !iso.Available() {
		if cfg.Sandbox.FailClosed {
			logger().Warnw("OS sandbox required but unavailable — tool execution will be blocked",
				"platform", status.Capabilities.Platform,
				"reason", iso.Reason(),
				"capabilities", status.Capabilities.Summary())
		} else {
			logger().Warnw("OS sandbox enabled but unavailable — proceeding without isolation",
				"platform", status.Capabilities.Platform,
				"reason", iso.Reason(),
				"capabilities", status.Capabilities.Summary())
		}
	} else {
		logger().Infow("OS sandbox initialized",
			"isolator", iso.Name(),
			"networkMode", cfg.Sandbox.NetworkMode,
			"failClosed", cfg.Sandbox.FailClosed)
	}

	return iso
}

