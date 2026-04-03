## ADDED Requirements

### Requirement: Sandbox executor nil-safety for P2P tool execution
The `Handler` MUST refuse tool execution when `sandboxExec` is nil. Both `handleToolInvoke` and `handleToolInvokePaid` MUST return `ResponseStatusDenied` with `ErrNoSandboxExecutor` instead of falling back to the in-process executor.

#### Scenario: Nil sandbox executor on free invocation
- **WHEN** a `tool_invoke` request arrives and `sandboxExec` is nil
- **THEN** the handler returns `{"status": "denied", "error": "tool execution refused: no sandbox executor configured for remote peer requests"}`
- **AND** the in-process executor is NOT called

#### Scenario: Nil sandbox executor on paid invocation
- **WHEN** a `tool_invoke_paid` request arrives and `sandboxExec` is nil
- **THEN** the handler returns `{"status": "denied"}` with `ErrNoSandboxExecutor`

### Requirement: P2P context injection
The `Handler` MUST call `ctxkeys.WithP2PRequest(ctx)` at the start of both `handleToolInvoke` and `handleToolInvokePaid` to mark the context as originating from a remote peer.

#### Scenario: P2P context propagated to tool execution
- **WHEN** a tool is invoked via P2P
- **THEN** `ctxkeys.IsP2PRequest(ctx)` returns true in all downstream tool handlers

### Requirement: SafetyLevel gate for P2P tool invocations
The `Handler` MUST check each tool's safety level against a configurable maximum before execution. Tools above the threshold MUST be rejected with `ErrToolSafetyBlocked`.

#### Scenario: Dangerous tool blocked at moderate threshold
- **WHEN** `maxSafetyLevel` is "moderate" and a Dangerous-level tool is invoked
- **THEN** the handler returns `{"status": "denied", "error": "tool blocked by P2P safety level policy"}`

#### Scenario: Whitelisted tool bypasses safety gate
- **WHEN** a tool is in the `allowedTools` list
- **THEN** it passes the safety gate regardless of its safety level

#### Scenario: No checker configured (backward compatible)
- **WHEN** no `SafetyLevelChecker` is set on the handler
- **THEN** all tools pass the safety gate

### Requirement: P2P safety configuration
`P2PConfig` MUST include `maxSafetyLevel` (string, default "moderate") and `allowedTools` (string slice, default empty) fields.

### Requirement: Sentinel errors
The protocol MUST define `ErrNoSandboxExecutor` and `ErrToolSafetyBlocked` sentinel errors.
