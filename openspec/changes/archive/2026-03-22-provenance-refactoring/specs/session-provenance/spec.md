## MODIFIED Requirements

### Requirement: Disabled provenance message
- **WHEN** any provenance command is run with provenance.enabled=false
- **THEN** the system displays an enable instruction message

The `status` command SHALL display the full configuration state followed by the enable instruction message, without blocking config display. All other provenance commands (`checkpoint`, `session`, `attribution`, `bundle`) SHALL display only the enable instruction message and return immediately.

#### Scenario: Disabled provenance message on data commands
- **WHEN** any provenance command other than `status` is run with `provenance.enabled=false`
- **THEN** the system prints "Provenance is disabled. Enable with: lango config set provenance.enabled true"
- **AND** the command returns without performing any data operation

#### Scenario: Disabled provenance status shows config
- **WHEN** `lango provenance status` is run with `provenance.enabled=false`
- **THEN** the system prints the provenance configuration (Enabled, Auto on Step Complete, etc.)
- **AND** appends "Provenance is disabled. Enable with: lango config set provenance.enabled true" at the end

### Requirement: Provenance Bundle Verification
The system SHALL support signed provenance bundles with redaction levels `none`, `content`, and `full`. The system SHALL reject invalid redaction levels with `ErrInvalidRedaction` in both `Export()` and `Verify()`.

#### Scenario: Export rejects invalid redaction
- **WHEN** `Export()` is called with a redaction level that is not `none`, `content`, or `full`
- **THEN** the system returns `ErrInvalidRedaction`
- **AND** no bundle is created

#### Scenario: Verify rejects invalid redaction in imported bundle
- **WHEN** `Verify()` is called on a bundle whose `redaction_level` field is not `none`, `content`, or `full`
- **THEN** the system returns `ErrInvalidRedaction`
- **AND** the bundle is not imported

#### Scenario: Remote peer verifies signed bundle
- **WHEN** a peer receives a provenance bundle over the provenance-specific P2P protocol
- **THEN** it verifies the bundle signature against the signer DID public key before import

#### Scenario: Tampered bundle rejected
- **WHEN** a signed bundle payload is modified after signing
- **THEN** verification fails and the bundle is rejected

#### Scenario: HTTP route rejects invalid redaction early
- **WHEN** a provenance push or fetch HTTP request contains an invalid redaction level
- **THEN** the route handler returns HTTP 400 Bad Request with a message indicating valid options
- **AND** no P2P connection is attempted

## ADDED Requirements

### Requirement: RedactionLevel validation method
The `RedactionLevel` type SHALL provide a `Valid()` method that returns `true` for `none`, `content`, and `full`, and `false` for all other values.

#### Scenario: Valid redaction levels
- **WHEN** `Valid()` is called on `RedactionNone`, `RedactionContent`, or `RedactionFull`
- **THEN** it returns `true`

#### Scenario: Invalid redaction levels
- **WHEN** `Valid()` is called on any other `RedactionLevel` value
- **THEN** it returns `false`
