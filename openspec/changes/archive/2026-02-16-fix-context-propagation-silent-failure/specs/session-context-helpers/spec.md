## ADDED Requirements

### Requirement: Shared session key context helpers
The `internal/session` package SHALL export `WithSessionKey(ctx, key)` and `SessionKeyFromContext(ctx)` functions to inject and extract session keys from `context.Context`. These helpers SHALL be the single canonical source for session key context propagation across all packages (`app`, `gateway`, channels).

#### Scenario: Inject and extract session key
- **WHEN** `session.WithSessionKey(ctx, "discord:123:456")` is called
- **THEN** `session.SessionKeyFromContext(resultCtx)` SHALL return `"discord:123:456"`

#### Scenario: Missing session key returns empty string
- **WHEN** `session.SessionKeyFromContext(context.Background())` is called on a context without a session key
- **THEN** the function SHALL return `""`

### Requirement: Remove duplicate helpers from app package
The `internal/app/tools.go` file SHALL NOT define `sessionKeyCtxKey`, `WithSessionKey`, or `SessionKeyFromContext`. All references within the `app` package SHALL use `session.WithSessionKey` and `session.SessionKeyFromContext` from `internal/session`.

#### Scenario: App package uses shared helpers
- **WHEN** the `wrapWithApproval` handler extracts the session key
- **THEN** it SHALL call `session.SessionKeyFromContext(ctx)`
