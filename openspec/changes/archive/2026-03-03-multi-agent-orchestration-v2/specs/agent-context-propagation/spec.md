## ADDED Requirements

### Requirement: Agent name context keys
The `ctxkeys` package SHALL provide `WithAgentName(ctx, name)` and `AgentNameFromContext(ctx)` functions for propagating agent identity through Go context.

#### Scenario: Set and retrieve agent name
- **WHEN** WithAgentName sets "operator" on a context
- **THEN** AgentNameFromContext SHALL return "operator"

#### Scenario: Missing agent name returns empty
- **WHEN** AgentNameFromContext is called on a context without agent name
- **THEN** it SHALL return an empty string

### Requirement: ADK tool adapter integration
The ADK tool adapter SHALL inject the current agent name into the Go context before tool execution, making it available to hooks and middleware.

#### Scenario: Agent name available in tool context
- **WHEN** a tool is executed via the ADK adapter within a sub-agent
- **THEN** the agent name SHALL be available via AgentNameFromContext in the tool's context
