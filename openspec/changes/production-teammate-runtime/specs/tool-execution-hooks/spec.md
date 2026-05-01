## MODIFIED Requirements

### Requirement: WithHooks middleware bridge

The package SHALL provide a `WithHooks(registry)` middleware that preserves structured blocked-call metadata before returning the existing blocked-tool error to the caller. When a pre-hook blocks execution, the runtime MUST be able to recover the structured tool name, blocking reason, and any dynamic allowlist metadata without reparsing the final error string. That structured metadata SHALL be exposed to capability-policy consumers through execution context or typed hook metadata, not by reparsing the returned error string.

#### Scenario: Structured blocked metadata survives hook block
- **WHEN** a pre-hook blocks a tool because of `DynamicAllowedTools`
- **THEN** `WithHooks` SHALL preserve the structured blocked-call metadata alongside the returned blocked-tool error

#### Scenario: Existing blocked-tool error surface is unchanged
- **WHEN** a caller receives a blocked-tool error from the middleware chain
- **THEN** the existing error contract SHALL remain intact for callers that only inspect the returned error
- **AND** structured metadata SHALL still be available to capability policy consumers
