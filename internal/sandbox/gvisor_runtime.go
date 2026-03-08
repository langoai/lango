package sandbox

import (
	"context"
)

// GVisorRuntime is a ContainerRuntime implementation backed by gVisor (runsc).
//
// gVisor provides user-space kernel isolation that is stronger than native
// process sandboxing but lighter than full VM-based containers. It intercepts
// application system calls via its Sentry component and services them without
// granting direct host-kernel access.
//
// This runtime is currently a stub. All methods behave as if gVisor is not
// installed: IsAvailable returns false and Run returns ErrRuntimeUnavailable.
// To enable gVisor support, install the runsc binary
// (see https://gvisor.dev/docs/user_guide/install/) and replace this stub
// with a real implementation that delegates to the runsc OCI runtime.
type GVisorRuntime struct{}

// Compile-time check: GVisorRuntime implements ContainerRuntime.
var _ ContainerRuntime = (*GVisorRuntime)(nil)

// NewGVisorRuntime creates a new GVisorRuntime stub. The returned runtime
// always reports as unavailable until a real gVisor integration is provided.
func NewGVisorRuntime() *GVisorRuntime {
	return &GVisorRuntime{}
}

// Run always returns ErrRuntimeUnavailable because gVisor support is not yet
// implemented. The cfg parameter is accepted but ignored.
func (r *GVisorRuntime) Run(_ context.Context, _ ContainerConfig) (*ExecutionResult, error) {
	return nil, ErrRuntimeUnavailable
}

// Cleanup is a no-op for the gVisor stub. It always returns nil because no
// containers are ever created.
func (r *GVisorRuntime) Cleanup(_ context.Context, _ string) error {
	return nil
}

// IsAvailable always returns false for the gVisor stub, indicating that the
// runsc binary is not present or not configured.
func (r *GVisorRuntime) IsAvailable(_ context.Context) bool {
	return false
}

// Name returns "gvisor", identifying this runtime in logs and probe chains.
func (r *GVisorRuntime) Name() string {
	return "gvisor"
}
