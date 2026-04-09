package os

// SandboxStatus summarizes the operational state of the sandbox subsystem.
// It combines configuration values with runtime probe results.
type SandboxStatus struct {
	// Enabled reflects the sandbox.enabled config value.
	Enabled bool
	// FailClosed reflects the sandbox.failClosed config value.
	FailClosed bool
	// Isolator is the active OS isolator (never nil).
	Isolator OSIsolator
	// Capabilities holds the platform primitive probe results.
	Capabilities PlatformCapabilities
}

// NewSandboxStatus creates a SandboxStatus from config and an isolator.
// If iso is nil (sandbox disabled), a disabledIsolator is substituted.
func NewSandboxStatus(enabled, failClosed bool, iso OSIsolator) SandboxStatus {
	if iso == nil {
		iso = &disabledIsolator{}
	}
	return SandboxStatus{
		Enabled:      enabled,
		FailClosed:   failClosed,
		Isolator:     iso,
		Capabilities: Probe(),
	}
}
