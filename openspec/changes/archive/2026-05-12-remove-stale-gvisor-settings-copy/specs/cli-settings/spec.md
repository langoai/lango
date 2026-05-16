## MODIFIED Requirements

### Requirement: P2P Sandbox settings form
The settings TUI SHALL provide a "P2P Sandbox" form with fields for tool isolation (enabled, timeout, max memory) and container sandbox (enabled, runtime, image, network mode, read-only rootfs, CPU quota, pool size, pool idle timeout). Container-specific fields SHALL only be visible when Container Sandbox is enabled.

#### Scenario: Container runtime description stays truth-aligned
- **WHEN** user opens the "P2P Sandbox" form
- **THEN** the container runtime field SHALL describe `auto` as detection-based
- **AND** SHALL describe `docker` as the preferred real container runtime
- **AND** SHALL describe `gvisor` as the current stub instead of implying that it is already the strongest available implementation
- **AND** SHALL describe `native` as the local fallback path
