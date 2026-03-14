package deadline

import "time"

// TimeoutConfig holds the config values needed for timeout resolution.
type TimeoutConfig struct {
	IdleTimeout       time.Duration
	RequestTimeout    time.Duration
	AutoExtendTimeout bool
	MaxRequestTimeout time.Duration
}

// ResolveTimeouts determines the idle timeout and hard ceiling from config.
//
// Precedence (highest to lowest):
//  1. IdleTimeout explicitly set  -> idle = IdleTimeout, ceiling = RequestTimeout (or 60m)
//  2. AutoExtendTimeout = true    -> idle = RequestTimeout, ceiling = MaxRequestTimeout (legacy compat)
//  3. Neither set                 -> idle = 0 (disabled), ceiling = RequestTimeout (fixed timeout)
//
// Returns (idleTimeout, hardCeiling). idleTimeout=0 means disabled (fixed timeout).
func ResolveTimeouts(cfg TimeoutConfig) (idleTimeout, hardCeiling time.Duration) {
	switch {
	case cfg.IdleTimeout > 0:
		// Explicit idle timeout configured.
		idleTimeout = cfg.IdleTimeout
		hardCeiling = cfg.RequestTimeout
		if hardCeiling <= 0 {
			hardCeiling = 60 * time.Minute
		}
		// Ensure ceiling > idle to be meaningful.
		if hardCeiling <= idleTimeout {
			hardCeiling = idleTimeout * 3
		}

	case cfg.IdleTimeout < 0:
		// Explicitly disabled — use fixed timeout.
		idleTimeout = 0
		hardCeiling = cfg.RequestTimeout
		if hardCeiling <= 0 {
			hardCeiling = 5 * time.Minute
		}

	case cfg.AutoExtendTimeout:
		// Legacy auto-extend mode: treat RequestTimeout as idle, MaxRequestTimeout as ceiling.
		idleTimeout = cfg.RequestTimeout
		if idleTimeout <= 0 {
			idleTimeout = 5 * time.Minute
		}
		hardCeiling = cfg.MaxRequestTimeout
		if hardCeiling <= 0 {
			hardCeiling = idleTimeout * 3
		}

	default:
		// No idle timeout — fixed RequestTimeout (backward compatible default).
		idleTimeout = 0
		hardCeiling = cfg.RequestTimeout
		if hardCeiling <= 0 {
			hardCeiling = 5 * time.Minute
		}
	}

	return idleTimeout, hardCeiling
}
