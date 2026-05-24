## MODIFIED Requirements

### Requirement: Policy types
The protected-path policy SHALL include the resolved application database path, resolved graph database path, and resolved envelope/keyfile paths after configuration is loaded.

#### Scenario: Runtime-resolved protected paths
- **WHEN** configuration load resolves custom session or graph database paths
- **THEN** the sandbox and protected-path denylist SHALL use those resolved paths rather than only default data-root assumptions

Subprocesses other than the storage broker child SHALL NOT inherit broker communication file descriptors.

#### Scenario: Non-broker subprocess launch
- **WHEN** the runtime launches a tool, MCP child, or skill subprocess that is not the storage broker
- **THEN** broker stdio/IPC file descriptors SHALL be closed-on-exec and unavailable to that child process
