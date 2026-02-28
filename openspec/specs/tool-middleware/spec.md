## Purpose

Composable middleware chain for cross-cutting tool concerns (learning observation, approval gating, browser recovery).

## Requirements

### Requirement: Middleware type
The system SHALL define a Middleware type as `func(tool *agent.Tool, next HandlerFunc) HandlerFunc` that wraps tool handlers.

#### Scenario: Middleware wraps handler
- **WHEN** a middleware is applied to a tool
- **THEN** it SHALL receive the tool metadata and next handler, returning a new handler

### Requirement: Chain applies middlewares in order
Chain SHALL apply middlewares so the first middleware is outermost (executed first).

#### Scenario: Two middlewares chain correctly
- **WHEN** middleware A and B are chained with Chain(tool, A, B)
- **THEN** execution order SHALL be: A's pre-logic -> B's pre-logic -> original handler -> B's post-logic -> A's post-logic

### Requirement: ChainAll applies to all tools
ChainAll SHALL apply the same middleware stack to every tool in the slice.

#### Scenario: ChainAll wraps all tools
- **WHEN** ChainAll is called with 3 tools and 2 middlewares
- **THEN** all 3 tools SHALL have both middlewares applied

### Requirement: WithLearning middleware
The WithLearning middleware SHALL call the learning observer after each tool execution with the tool name, params, result, and error.

#### Scenario: Learning observes tool result
- **WHEN** a tool wrapped with WithLearning executes
- **THEN** observer.OnToolResult SHALL be called with session key, tool name, params, result, and error

### Requirement: WithApproval middleware
The WithApproval middleware SHALL gate tool execution behind an approval flow based on configured policy.

#### Scenario: Dangerous tool requires approval
- **WHEN** a tool with dangerous safety level is executed under "dangerous" policy
- **THEN** the approval provider SHALL be consulted before execution

#### Scenario: Exempt tool bypasses approval
- **WHEN** a tool listed in ExemptTools is executed
- **THEN** execution SHALL proceed without approval

### Requirement: WithBrowserRecovery middleware
The WithBrowserRecovery middleware SHALL recover from panics in browser tool handlers and retry once on ErrBrowserPanic.

#### Scenario: Browser panic triggers retry
- **WHEN** a browser tool panics with ErrBrowserPanic
- **THEN** the session SHALL be closed and the handler retried once
