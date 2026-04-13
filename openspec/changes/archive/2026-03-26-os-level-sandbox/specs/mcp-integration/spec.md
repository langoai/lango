## ADDED Requirements

### Requirement: MCP stdio server OS sandbox
The MCP `ServerConnection` SHALL support optional OS-level sandbox for stdio server processes via `SetOSIsolator()`, applied at transport creation time with `MCPServerPolicy()` (network=allow, filesystem restricted).

#### Scenario: Stdio server sandboxed
- **WHEN** an MCP stdio server is started with isolator configured
- **THEN** the server process SHALL run with filesystem restrictions (read-global, write-/tmp only) while retaining network access

#### Scenario: Sandbox error is non-fatal
- **WHEN** the isolator returns an error during transport creation
- **THEN** the server SHALL start without sandbox and log a warning
