## ADDED Requirements

### Requirement: Session sentinel errors
The system SHALL define `ErrSessionNotFound` and `ErrDuplicateSession` in `session/errors.go`.

#### Scenario: Replace string matching with errors.Is
- **WHEN** `adk/session_service.go` checks for "session not found" errors
- **THEN** it SHALL use `errors.Is(err, session.ErrSessionNotFound)` instead of `strings.Contains`

#### Scenario: Replace UNIQUE constraint matching
- **WHEN** `adk/session_service.go` checks for duplicate session errors
- **THEN** it SHALL use `errors.Is(err, session.ErrDuplicateSession)` instead of string matching

### Requirement: Gateway sentinel errors
The system SHALL define `ErrNoCompanion`, `ErrApprovalTimeout`, `ErrAgentNotReady` in `gateway/errors.go`.

#### Scenario: Gateway error handling
- **WHEN** gateway operations encounter known error conditions
- **THEN** they SHALL return sentinel errors instead of ad-hoc error messages

### Requirement: Workflow sentinel errors
The system SHALL define `ErrWorkflowNameEmpty`, `ErrNoWorkflowSteps`, `ErrStepIDEmpty` in `workflow/errors.go`.

#### Scenario: Workflow validation errors
- **WHEN** workflow validation fails
- **THEN** it SHALL return sentinel errors for programmatic handling

### Requirement: Knowledge sentinel errors
The system SHALL define `ErrKnowledgeNotFound`, `ErrLearningNotFound` in `knowledge/errors.go`.

#### Scenario: Knowledge lookup errors
- **WHEN** knowledge or learning lookups find no results
- **THEN** they SHALL return sentinel errors

### Requirement: Security sentinel errors
The system SHALL define `ErrKeyNotFound`, `ErrNoEncryptionKeys`, `ErrDecryptionFailed` in `security/errors.go`.

#### Scenario: Security operation errors
- **WHEN** security operations encounter known failure modes
- **THEN** they SHALL return sentinel errors

### Requirement: Gateway RPCError type
The system SHALL define `RPCError` struct with `Code int` and `Message string` fields implementing the `error` interface in `gateway/errors.go`.

#### Scenario: Structured RPC errors
- **WHEN** `gateway/server.go` creates RPC error responses
- **THEN** it SHALL use the named `RPCError` type instead of anonymous structs
