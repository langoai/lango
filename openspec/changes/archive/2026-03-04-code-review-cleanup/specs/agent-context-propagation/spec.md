## MODIFIED Requirements

### Requirement: Agent name context keys
The `ctxkeys` package SHALL provide `WithAgentName(ctx, name)` and `AgentNameFromContext(ctx)` functions for propagating agent identity through Go context. The `toolchain` package SHALL delegate its `WithAgentName` and `AgentNameFromContext` functions to the `ctxkeys` canonical implementations, ensuring a single context key is used across the entire codebase.

#### Scenario: Set and retrieve agent name
- **WHEN** WithAgentName sets "operator" on a context
- **THEN** AgentNameFromContext SHALL return "operator"

#### Scenario: Missing agent name returns empty
- **WHEN** AgentNameFromContext is called on a context without agent name
- **THEN** it SHALL return an empty string

#### Scenario: toolchain delegates to ctxkeys
- **WHEN** `toolchain.WithAgentName` sets a name on a context
- **THEN** `ctxkeys.AgentNameFromContext` SHALL return the same name (single canonical key)

#### Scenario: Cross-package context key compatibility
- **WHEN** `ctxkeys.WithAgentName` sets a name on a context
- **THEN** `toolchain.AgentNameFromContext` SHALL return the same name
