## Purpose

Unified status dashboard command (`lango status`) that combines health, configuration state, active channels, and feature status into a single view. Replaces the need to run health/doctor/metrics separately.
## Requirements
### Requirement: Unified status dashboard command
The system SHALL provide a `lango status` command that displays system health, configuration state, active channels, and feature status in a single dashboard view. When `--addr` is omitted, the live server probe SHALL use the gateway address resolved from configured `server.host` and `server.port`, falling back to `http://localhost:18789` only when configuration is unavailable, blank, or zero-valued.

#### Scenario: Status with server not running
- **WHEN** user runs `lango status` and the server is not running
- **THEN** system displays config-based status (profile, gateway address, provider, features) with server marked as "not running"

#### Scenario: Status with server running
- **WHEN** user runs `lango status` and the server is running
- **THEN** system displays live health data alongside config-based status with server marked as "running"

#### Scenario: JSON output
- **WHEN** user runs `lango status --output json`
- **THEN** system outputs all status data as a JSON object with version, profile, serverUp, gateway, provider, model, features, channels, and serverInfo fields

#### Scenario: Status probe uses configured gateway by default
- **WHEN** configuration sets `server.host` and `server.port`
- **AND** the user runs `lango status` without `--addr`
- **THEN** the command SHALL probe `/health` on that configured gateway address
- **AND** the displayed gateway field SHALL match the same configured gateway address

#### Scenario: Status explicit address override
- **WHEN** the user runs `lango status --addr <url>`
- **THEN** the command SHALL probe `/health` on `<url>` instead of the configured gateway address

#### Scenario: Root status rejects an unknown output format before bootstrap
- **WHEN** the operator runs `lango status --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the bootstrap loader

#### Scenario: Dead-letter status rejects an unknown output format before bridge loading
- **WHEN** the operator runs `lango status dead-letters --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the dead-letter bridge loader

#### Scenario: Root status table output uses the command writer
- **WHEN** `lango status` renders the human-readable dashboard
- **THEN** it SHALL write the dashboard through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the full dashboard output

#### Scenario: Root status version text stays plain and single-line
- **WHEN** injected build version text contains ANSI/OSC escape sequences or embedded newlines before root status serialization
- **THEN** the status command SHALL strip those control sequences
- **AND** it SHALL normalize the stored top-level `version` field to a single line before table or JSON output

#### Scenario: Rendered status CLI text stays plain and single-line
- **WHEN** provider/model labels, feature names/details, channel names, or dead-letter labels contain ANSI/OSC escape sequences or embedded newlines
- **THEN** the status CLI SHALL strip those control sequences
- **AND** it SHALL normalize the displayed text to a single line before rendering it

#### Scenario: Collected status model text is replay-safe
- **WHEN** provider/model labels, feature names/details, channel names, or live feature reasons contain ANSI/OSC escape sequences or embedded newlines before status collection or enrichment
- **THEN** the status command SHALL strip those control sequences
- **AND** it SHALL normalize the stored status model text to a single line before replay or JSON output

#### Scenario: Live server feature model text is replay-safe
- **WHEN** live `/health` feature names, reasons, or suggestions contain ANSI/OSC escape sequences or embedded newlines before status collection
- **THEN** the status command SHALL strip those control sequences
- **AND** it SHALL normalize the stored `serverInfo.features` text to a single line before replay or JSON output

#### Scenario: Dead-letter retry result model text is replay-safe
- **WHEN** dead-letter retry message, follow-up error text, subtype/family/reason/dispatch labels, or background-task status text contain ANSI/OSC escape sequences or embedded newlines before result construction
- **THEN** the status command SHALL strip those control sequences
- **AND** it SHALL normalize the stored retry-result text to a single line before replay or JSON output

#### Scenario: JSON error payload stays plain and single-line
- **WHEN** `lango status` emits a JSON error payload and the underlying error text contains ANSI/OSC escape sequences or embedded newlines
- **THEN** the status command SHALL strip those control sequences
- **AND** it SHALL normalize the serialized `error` field to a single line before JSON output

#### Scenario: Dead-letter summary model text is replay-safe
- **WHEN** dead-letter summary bucket labels, top latest reason labels, top latest actor labels, or top latest dispatch-reference labels contain ANSI/OSC escape sequences or embedded newlines before summary aggregation
- **THEN** the status command SHALL strip those control sequences
- **AND** it SHALL normalize the stored summary-model text to a single line before replay or JSON output

#### Scenario: Dead-letter list/detail JSON payload text is replay-safe
- **WHEN** dead-letter backlog JSON fields or dead-letter detail JSON fields contain ANSI/OSC escape sequences or embedded newlines before CLI serialization
- **THEN** the status command SHALL strip those control sequences
- **AND** it SHALL normalize the serialized JSON payload text to a single line while preserving the existing data shape

#### Scenario: Dead-letter retry prompt and non-JSON error text stay plain and single-line
- **WHEN** a transaction receipt ID contains ANSI/OSC escape sequences or embedded newlines before dead-letter retry confirmation or non-JSON error reporting
- **THEN** the status command SHALL strip those control sequences from the displayed transaction label
- **AND** it SHALL normalize the prompt and error text to a single line while preserving the lookup input passed to the retry flow

#### Scenario: Dead-letter invalid-flag validation errors stay plain and single-line
- **WHEN** invalid dead-letter subtype, family, or RFC3339 flag values contain ANSI/OSC escape sequences or embedded newlines before CLI validation fails
- **THEN** the status command SHALL strip those control sequences from the echoed invalid value
- **AND** it SHALL normalize the non-JSON validation error text to a single line

#### Scenario: Non-JSON status command errors stay plain and single-line
- **WHEN** any downstream status command error contains ANSI/OSC escape sequences or embedded newlines before non-JSON CLI handling
- **THEN** the status command SHALL strip those control sequences at the shared non-JSON error boundary
- **AND** it SHALL normalize the surfaced error text to a single line
- **AND** it SHALL preserve the original error as the unwrap cause

### Requirement: Feature collection from config
The system SHALL collect feature status for 14 features: Knowledge, Embedding & RAG, Graph, Obs. Memory, Librarian, Multi-Agent, Cron, Background, Workflow, MCP, P2P, Payment, Economy, A2A.

#### Scenario: All features disabled
- **WHEN** default config is used
- **THEN** all optional features report as disabled

#### Scenario: MCP detail shows server count
- **WHEN** MCP is enabled with 2 servers configured
- **THEN** MCP feature detail shows "2 server(s)"

### Requirement: Channel collection
The system SHALL list active channels (telegram, discord, slack) based on their Enabled config flag.

#### Scenario: Multiple channels enabled
- **WHEN** Telegram and Slack are enabled in config
- **THEN** channels list contains "telegram" and "slack"

### Requirement: Provenance Feature Line
The status dashboard SHALL include a "Provenance" feature line showing whether provenance is enabled.

#### Scenario: Provenance enabled
- **WHEN** provenance.enabled is true
- **THEN** the status dashboard shows Provenance as enabled

### Requirement: Status command shows context profile
`lango status` SHALL display the active context profile name alongside existing feature information. If no profile is set, the profile field SHALL show "none" or be omitted.

#### Scenario: Profile shown in status output
- **WHEN** `contextProfile: balanced` is set
- **THEN** `lango status` output includes "Profile: balanced" in the dashboard header or feature section

### Requirement: Feature detail includes reason
The `Detail` field of context-related features in `collectFeatures()` SHALL reflect the `FeatureStatus.Reason` when available, providing users actionable context about why a feature is disabled.

#### Scenario: Embedding detail shows reason
- **WHEN** embedding is disabled because no provider is configured
- **THEN** status output for "Embedding & RAG" shows `Detail: "no provider configured"` instead of empty string

#### Scenario: Enabled feature detail unchanged
- **WHEN** knowledge is enabled and healthy
- **THEN** status output for "Knowledge" shows existing detail behavior (empty or provider info)

### Requirement: Dead-letter retry confirmation uses shared command streams
`lango status dead-letter retry` SHALL drive its interactive confirmation through the shared confirmation helper using Cobra command input/output streams. Empty input or EOF SHALL continue to abort the retry without invoking the retry mutation.

#### Scenario: Dead-letter retry denial aborts cleanly
- **WHEN** `lango status dead-letter retry <id>` is run and the operator answers `n`
- **THEN** the command SHALL print an aborted message
- **AND** it SHALL NOT invoke the retry mutation

#### Scenario: Dead-letter retry EOF aborts cleanly
- **WHEN** `lango status dead-letter retry <id>` is run and the confirmation input reaches EOF without approval
- **THEN** the command SHALL abort the retry without invoking the retry mutation

#### Scenario: Dead-letter retry prompt uses command output
- **WHEN** `lango status dead-letter retry <id>` prompts for confirmation
- **THEN** the sanitized transaction receipt ID and `[y/N]:` suffix SHALL be written through the Cobra command output stream

### Requirement: Dead-letter retry EOF uses shared deny behavior
`lango status dead-letter retry` SHALL use the shared EOF-deny confirmation helper for its default confirmation path.

#### Scenario: Dead-letter retry EOF aborts cleanly through shared helper
- **WHEN** the retry confirmation reaches EOF before approval
- **THEN** the command SHALL abort the retry without invoking the mutation path
