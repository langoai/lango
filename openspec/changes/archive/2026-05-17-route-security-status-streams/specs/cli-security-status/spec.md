## MODIFIED Requirements

### Requirement: Security status output routing
`lango security status` SHALL route human-readable and JSON output through the
Cobra command writer instead of writing directly to process stdout. Status
diagnostics and warnings SHALL route through the Cobra command error writer
instead of process-global stderr.

#### Scenario: Security status output uses the command writer
- **WHEN** `lango security status` renders table or JSON output
- **THEN** the command SHALL write the full output through the Cobra command
  output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the
  command output

#### Scenario: Security status warning uses command error writer
- **WHEN** the non-interactive status path emits a passphrase acquisition warning
- **THEN** the warning SHALL be written through the Cobra command error writer
- **AND** it SHALL NOT require intercepting process-global stderr

#### Scenario: Security status rejects unknown output before bootstrap
- **WHEN** `lango security status --output yaml` is run
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke bootstrap-dependent work

### Requirement: Non-interactive mini-bootstrap for status

The system SHALL provide a `readDBStatusNonInteractive` helper that runs a
minimal bootstrap (envelope load → non-interactive passphrase → MK unwrap →
read-only DB open → read counts → close) without triggering interactive prompts
or schema migration. The helper SHALL handle both envelope-based and legacy
installations.

#### Scenario: Keyring provider passed to non-interactive acquisition
- **WHEN** `readDBStatusNonInteractive` acquires a passphrase
- **THEN** it SHALL pass the status secure-provider detector result as the
  `KeyringProvider` option

#### Scenario: Broker startup is injectable for status reads
- **WHEN** `readDBStatusNonInteractive` needs broker-backed DB counts
- **THEN** it SHALL start the broker through a replaceable status broker starter
- **AND** tests SHALL be able to verify the `DBStatusSummaryRequest` without
  launching a real broker process

#### Scenario: Broker startup failure degrades
- **WHEN** the status broker starter returns an error
- **THEN** `readDBStatusNonInteractive` SHALL return a zero-valued
  `dbStatusResult`
- **AND** it SHALL NOT panic or emit misleading DB counts

#### Scenario: No passphrase available
- **WHEN** `readDBStatusNonInteractive` is called and `AcquireNonInteractive`
  returns an error
- **THEN** the helper returns a zero-valued `dbStatusResult` (all counts 0)
- **AND** no DB open attempt is made
- **AND** non-`ErrNoNonInteractiveSource` errors SHALL be capturable through the
  supplied warning writer
