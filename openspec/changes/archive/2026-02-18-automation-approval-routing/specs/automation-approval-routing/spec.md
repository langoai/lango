## ADDED Requirements

### Requirement: Approval target context helpers
The system SHALL provide `WithApprovalTarget(ctx, target)` and `ApprovalTargetFromContext(ctx)` functions in the `approval` package for injecting and retrieving an explicit approval routing target from a context.

#### Scenario: Set and retrieve approval target
- **WHEN** `WithApprovalTarget(ctx, "telegram:123456789")` is called
- **THEN** `ApprovalTargetFromContext(ctx)` SHALL return `"telegram:123456789"`

#### Scenario: No approval target set
- **WHEN** no approval target has been set on the context
- **THEN** `ApprovalTargetFromContext(ctx)` SHALL return an empty string

### Requirement: Approval target overrides session key in wrapWithApproval
The `wrapWithApproval` function SHALL use the approval target from context as the session key when available, falling back to the standard session key from context when no approval target is set.

#### Scenario: Approval target present
- **WHEN** a tool requires approval
- **AND** the context has an approval target of `"telegram:123456789"`
- **THEN** the approval request SHALL use `"telegram:123456789"` as the session key
- **AND** the Telegram approval provider SHALL handle the request

#### Scenario: No approval target (backward compatible)
- **WHEN** a tool requires approval
- **AND** no approval target is set on the context
- **THEN** the approval request SHALL use the standard session key from context
- **AND** routing behavior SHALL be identical to before this change

### Requirement: Cron executor injects approval target from DeliverTo
The cron executor SHALL inject the first `DeliverTo` entry as the approval target when it contains a colon separator (channel:id format).

#### Scenario: DeliverTo with channel ID
- **WHEN** a cron job has `DeliverTo` of `["telegram:123456789"]`
- **THEN** the execution context SHALL have approval target `"telegram:123456789"`
- **AND** tool approval requests SHALL route to Telegram chat 123456789

#### Scenario: DeliverTo with bare channel name
- **WHEN** a cron job has `DeliverTo` of `["telegram"]`
- **THEN** no approval target SHALL be injected
- **AND** the system SHALL fall back to default routing (TTY)

#### Scenario: Empty DeliverTo
- **WHEN** a cron job has empty `DeliverTo`
- **THEN** no approval target SHALL be injected

### Requirement: Background manager injects approval target from origin
The background manager SHALL inject the task's origin session or origin channel as the approval target, preferring `OriginSession` over `OriginChannel`.

#### Scenario: Task with OriginSession
- **WHEN** a background task has `OriginSession` of `"telegram:123:456"`
- **THEN** the execution context SHALL have approval target `"telegram:123:456"`

#### Scenario: Task with OriginChannel only
- **WHEN** a background task has empty `OriginSession`
- **AND** `OriginChannel` is `"telegram:123456789"`
- **THEN** the execution context SHALL have approval target `"telegram:123456789"`

#### Scenario: Task with bare OriginChannel
- **WHEN** a background task has empty `OriginSession`
- **AND** `OriginChannel` is `"telegram"` (no colon)
- **THEN** no approval target SHALL be injected

#### Scenario: Task with no origin info
- **WHEN** a background task has empty `OriginSession` and empty `OriginChannel`
- **THEN** no approval target SHALL be injected
