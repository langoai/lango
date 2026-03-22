## MODIFIED Requirements

### Requirement: Provenance CLI
The system SHALL provide working `lango provenance` CLI commands for checkpoint, session tree, attribution, and bundle import/export operations.

#### Scenario: Session tree command returns persisted subtree
- **WHEN** user runs `lango provenance session tree <session-key> --depth <n>`
- **THEN** the CLI loads the subtree from the persistent session provenance store
- **AND** it prints the root plus descendants up to the requested depth

#### Scenario: Session list command returns persisted nodes
- **WHEN** user runs `lango provenance session list --limit <n> --status <status>`
- **THEN** the CLI returns persisted session nodes ordered by `created_at` descending
- **AND** status filtering is applied when provided

#### Scenario: Attribution show returns raw provenance rows
- **WHEN** user runs `lango provenance attribution show <session-key>`
- **THEN** the CLI returns raw attribution rows for the session
- **AND** it includes joined token summaries when available

#### Scenario: Attribution report returns aggregated view
- **WHEN** user runs `lango provenance attribution report <session-key>`
- **THEN** the CLI returns `by_author`, `by_file`, `total_tokens`, and checkpoint count aggregates

#### Scenario: Bundle export returns signed redacted bundle
- **WHEN** user runs `lango provenance bundle export <session-key> --redaction <level>`
- **THEN** the CLI emits a provenance bundle for that session
- **AND** the bundle includes `signer_did`, `signature_algorithm`, and `signature`
- **AND** the bundle applies the requested redaction level

#### Scenario: Bundle import verifies and stores only
- **WHEN** user runs `lango provenance bundle import <file>`
- **THEN** the CLI verifies signer DID and signature before import
- **AND** imported data is stored only in provenance-owned persistence
- **AND** existing run/session/workspace state is not mutated

### Requirement: Session Tree Tracking
The system SHALL persist session hierarchy in the `SessionProvenance` Ent store and SHALL update it from runtime child-session lifecycle events.

#### Scenario: Runtime fork persists child session node
- **WHEN** a child session is forked at runtime
- **THEN** a session provenance node is persisted with the parent relationship and agent name

#### Scenario: Runtime merge persists closed session state
- **WHEN** a child session is merged at runtime
- **THEN** the persisted node status becomes `merged`
- **AND** `closed_at` is set

#### Scenario: Runtime discard persists closed session state
- **WHEN** a child session is discarded at runtime
- **THEN** the persisted node status becomes `discarded`
- **AND** `closed_at` is set

## ADDED Requirements

### Requirement: Attribution Tracking
The system SHALL persist git-aware attribution records for provenance reporting.

#### Scenario: Workspace merge creates attribution records
- **WHEN** a workspace task branch is merged
- **THEN** attribution rows are recorded with workspace id, author identity, commit hash, file path, and line deltas

#### Scenario: Workspace bundle apply creates attribution records
- **WHEN** a workspace bundle is applied
- **THEN** attribution rows are recorded for the imported commit/file evidence

#### Scenario: Token-only attribution report for non-workspace sessions
- **WHEN** a session has token usage but no workspace git evidence
- **THEN** attribution reporting still succeeds
- **AND** file and commit sections are empty while author/session token totals are populated

### Requirement: Provenance Bundle Verification
The system SHALL support signed provenance bundles that are verifiable by remote peers using DID/public-key identity.

#### Scenario: Remote peer verifies signed bundle
- **WHEN** a peer receives a provenance bundle over P2P
- **THEN** it verifies that the signature matches the canonical payload and signer DID public key before accepting the bundle

#### Scenario: Tampered bundle rejected
- **WHEN** the bundle payload is modified after signing
- **THEN** verification fails
- **AND** the bundle is rejected

### Requirement: Provenance P2P Transport
The system SHALL provide a dedicated provenance transport for bundle exchange independent of workspace git bundle transport.

#### Scenario: Provenance bundle transferred over dedicated protocol
- **WHEN** a signed provenance bundle is sent to a remote peer
- **THEN** it is exchanged over the provenance-specific P2P protocol
- **AND** the receiving side routes it through the provenance import verification path
