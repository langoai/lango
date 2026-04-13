package sandbox

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/config"
)

// ContainerExecutor runs tool invocations through a container runtime.
// It probes available runtimes in priority order and falls back to native
// subprocess execution when no container runtime is available.
type ContainerExecutor struct {
	runtime          ContainerRuntime
	cfg              Config
	image            string
	networkMode      string
	readOnly         bool
	cpuQuotaUS       int64
	requireContainer bool
}

// NewContainerExecutor creates a ContainerExecutor by probing runtimes in order.
// Priority: docker (if requested or auto) > gvisor (if requested or auto) > native.
// When requireContainer is true and no container runtime is available, an error
// is returned instead of silently falling back to NativeRuntime.
func NewContainerExecutor(cfg Config, containerCfg config.ContainerSandboxConfig) (*ContainerExecutor, error) {
	ctx := context.Background()
	runtimeName := containerCfg.Runtime

	readOnly := true
	if containerCfg.ReadOnlyRootfs != nil {
		readOnly = *containerCfg.ReadOnlyRootfs
	}

	exec := &ContainerExecutor{
		cfg:              cfg,
		image:            containerCfg.Image,
		networkMode:      containerCfg.NetworkMode,
		readOnly:         readOnly,
		cpuQuotaUS:       containerCfg.CPUQuotaUS,
		requireContainer: containerCfg.RequireContainer,
	}

	// Try Docker runtime.
	if runtimeName == "docker" || runtimeName == "auto" {
		dr, err := NewDockerRuntime()
		if err == nil && dr.IsAvailable(ctx) {
			exec.runtime = dr
			return exec, nil
		}
		if runtimeName == "docker" {
			return nil, fmt.Errorf("docker runtime requested but unavailable: %w", ErrRuntimeUnavailable)
		}
	}

	// Try gVisor runtime.
	if runtimeName == "gvisor" || runtimeName == "auto" {
		gr := NewGVisorRuntime()
		if gr.IsAvailable(ctx) {
			exec.runtime = gr
			return exec, nil
		}
		if runtimeName == "gvisor" {
			return nil, fmt.Errorf("gvisor runtime requested but unavailable: %w", ErrRuntimeUnavailable)
		}
	}

	// Fail-closed: refuse to proceed when container isolation is required
	// but no container runtime is available.
	if exec.requireContainer {
		return nil, fmt.Errorf(
			"container runtime required but unavailable: %w",
			ErrRuntimeUnavailable,
		)
	}

	// Fallback to native (subprocess).
	exec.runtime = NewNativeRuntime(cfg)
	return exec, nil
}

// Execute runs a tool through the container runtime.
func (e *ContainerExecutor) Execute(ctx context.Context, toolName string, params map[string]interface{}) (map[string]interface{}, error) {
	ccfg := ContainerConfig{
		Image:          e.image,
		ToolName:       toolName,
		NetworkMode:    e.networkMode,
		Params:         params,
		MemoryLimitMB:  int64(e.cfg.MaxMemoryMB),
		CPUQuotaUS:     e.cpuQuotaUS,
		ReadOnlyRootfs: e.readOnly,
		Timeout:        e.cfg.TimeoutPerTool,
	}

	result, err := e.runtime.Run(ctx, ccfg)
	if err != nil {
		return nil, err
	}

	return result.Output, nil
}

// RuntimeName returns the name of the active container runtime.
func (e *ContainerExecutor) RuntimeName() string {
	return e.runtime.Name()
}

// Runtime returns the underlying ContainerRuntime for advanced operations.
func (e *ContainerExecutor) Runtime() ContainerRuntime {
	return e.runtime
}
