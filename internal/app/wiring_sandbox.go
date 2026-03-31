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
	if !iso.Available() {
		info := sandboxos.Probe()
		if cfg.Sandbox.FailClosed {
			logger().Warnw("OS sandbox required but unavailable — tool execution will be blocked",
				"platform", info.Platform,
				"capabilities", info.Summary())
		} else {
			logger().Warnw("OS sandbox enabled but unavailable on this platform — proceeding without isolation",
				"platform", info.Platform)
		}
	} else {
		logger().Infow("OS sandbox initialized",
			"isolator", iso.Name(),
			"networkMode", cfg.Sandbox.NetworkMode,
			"failClosed", cfg.Sandbox.FailClosed)
	}

	return iso
}

