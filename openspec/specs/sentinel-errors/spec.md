## Purpose

Capability spec for sentinel-errors. See requirements below for scope and behavior contracts.

## Requirements

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

### Requirement: Protocol sentinel errors
The system SHALL define sentinel errors in `protocol/messages.go` for common P2P protocol error conditions: `ErrMissingToolName`, `ErrAgentCardUnavailable`, `ErrNoApprovalHandler`, `ErrDeniedByOwner`, `ErrExecutorNotConfigured`, `ErrInvalidSession`, `ErrInvalidPaymentAuth`.

#### Scenario: Handler uses sentinel errors
- **WHEN** the protocol handler encounters a known error condition (missing tool name, no card, no approval handler, denied by owner, no executor, invalid session, invalid payment)
- **THEN** it SHALL use the sentinel error's `.Error()` message in the response Error field

#### Scenario: Sentinel errors are matchable
- **WHEN** a caller receives a protocol error
- **THEN** it SHALL be able to use `errors.Is()` to match against the sentinel errors

### Requirement: Firewall sentinel errors
The system SHALL define sentinel errors in `firewall/firewall.go`: `ErrRateLimitExceeded`, `ErrGlobalRateLimitExceeded`, `ErrQueryDenied`, `ErrNoMatchingAllowRule`.

#### Scenario: Rate limit errors wrap sentinel
- **WHEN** a peer exceeds the rate limit
- **THEN** `FilterQuery` SHALL return an error wrapping `ErrRateLimitExceeded` with `%w`

#### Scenario: ACL deny errors wrap sentinel
- **WHEN** a firewall deny rule matches
- **THEN** `FilterQuery` SHALL return an error wrapping `ErrQueryDenied`

#### Scenario: No matching allow rule wraps sentinel
- **WHEN** no allow rule matches and default-deny applies
- **THEN** `FilterQuery` SHALL return an error wrapping `ErrNoMatchingAllowRule`

### Requirement: ZKP unsupported scheme error
The system SHALL define `ErrUnsupportedScheme` in `zkp/zkp.go`.

#### Scenario: Unknown scheme returns sentinel
- **WHEN** a ZKP operation encounters an unknown proving scheme
- **THEN** it SHALL return an error wrapping `ErrUnsupportedScheme`

### Requirement: Session expiry sentinel error
The system SHALL define `ErrSessionExpired` in `session/errors.go` alongside existing session sentinel errors.

#### Scenario: EntStore wraps TTL expiry with ErrSessionExpired
- **WHEN** `EntStore.Get()` finds a session whose `UpdatedAt` exceeds the configured TTL
- **THEN** it SHALL return an error wrapping `ErrSessionExpired` using `fmt.Errorf("get session %q: %w", key, ErrSessionExpired)`

#### Scenario: ErrSessionExpired is matchable via errors.Is
- **WHEN** a caller receives a TTL expiry error from `EntStore.Get()`
- **THEN** `errors.Is(err, ErrSessionExpired)` SHALL return `true`

### Requirement: Alert metadata structure
The sentinel `Alert` type SHALL use a typed `AlertMetadata` struct instead of `map[string]interface{}` for the `Metadata` field. The struct SHALL include fields: `Count`, `Window`, `Amount`, `Threshold`, `Elapsed`, `PreviousBalance`, `NewBalance` with `json:"...,omitempty"` tags.

#### Scenario: Detector populates typed metadata
- **WHEN** `RapidCreationDetector` generates an alert for excessive deal creation
- **THEN** the alert's `Metadata.Count` and `Metadata.Window` fields are populated as typed values

#### Scenario: JSON marshaling preserves omitempty
- **WHEN** an alert has only `Count` and `Window` set in metadata
- **THEN** JSON output omits `amount`, `threshold`, `elapsed`, `previousBalance`, `newBalance` fields

### Requirement: Shared window counter for detectors
The sentinel package SHALL provide a `windowCounter` struct that encapsulates sliding-window counting logic shared by `RapidCreationDetector` and `RepeatedDisputeDetector`.

#### Scenario: windowCounter records and counts
- **WHEN** `record(key)` is called multiple times within the configured window
- **THEN** it returns the count of entries within the window, pruning expired entries

#### Scenario: Detectors embed windowCounter
- **WHEN** `RapidCreationDetector` and `RepeatedDisputeDetector` are initialized
- **THEN** both embed `windowCounter` instead of duplicating sliding-window logic

### Requirement: Domain error sentinels for escrow
The `internal/economy/escrow/` package SHALL export `ErrNotFunded` and `ErrInvalidStatus` sentinel errors.

#### Scenario: ErrNotFunded matching
- **WHEN** an escrow operation fails because funds are not deposited
- **THEN** the returned error wraps `ErrNotFunded` and is matchable via `errors.Is`

### Requirement: Domain error sentinel for workspace
The `internal/p2p/workspace/` package SHALL export `ErrWorkspaceNotFound`.

#### Scenario: Workspace lookup failure
- **WHEN** `manager.GetWorkspace(id)` is called with a non-existent workspace ID
- **THEN** the returned error wraps `ErrWorkspaceNotFound`

### Requirement: Domain error sentinel for workflow
The `internal/cli/workflow/` package SHALL export `ErrWorkflowDisabled`.

#### Scenario: Workflow operations when disabled
- **WHEN** any workflow function is called while the workflow engine is not enabled
- **THEN** it returns `ErrWorkflowDisabled`
