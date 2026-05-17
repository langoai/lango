## Purpose

Capability spec for cli-reference. See requirements below for scope and behavior contracts.
## Requirements
### Requirement: Security extension commands documented in CLI reference
The docs/cli/index.md SHALL include keyring (store/clear/status), db-migrate, db-decrypt, and kms (status/test/keys) commands in the Security table.

#### Scenario: Security table contains all 13 commands
- **WHEN** a user reads docs/cli/index.md Security section
- **THEN** the table SHALL list 13 security commands including the 8 new extension commands

### Requirement: P2P Network section in CLI reference
The docs/cli/index.md SHALL include a P2P Network table with all 17 P2P commands (status, peers, connect, disconnect, firewall, discover, identity, reputation, pricing, session, sandbox).

#### Scenario: P2P table exists between Payment and Automation
- **WHEN** a user reads docs/cli/index.md
- **THEN** a "P2P Network" section SHALL appear with 17 command entries

### Requirement: Background task commands in CLI reference
The docs/cli/index.md Automation section SHALL include bg list, bg status, bg cancel, and bg result commands.

#### Scenario: bg commands appear in Automation table
- **WHEN** a user reads the Automation section of docs/cli/index.md
- **THEN** 4 bg commands SHALL be listed after the workflow commands

### Requirement: README CLI section includes all commands
The README.md CLI Commands section SHALL include security keyring/db/kms commands, p2p session/sandbox commands, and bg commands.

#### Scenario: README CLI section is complete
- **WHEN** a user reads README.md CLI Commands section
- **THEN** all security extension, p2p session/sandbox, and bg commands SHALL be listed

### Requirement: Economy commands in CLI reference
The docs/cli/index.md SHALL include an Economy section with a table listing all 5 economy commands: `lango economy budget status`, `lango economy risk status`, `lango economy pricing status`, `lango economy negotiate status`, and `lango economy escrow status`.

#### Scenario: Economy table exists in CLI index
- **WHEN** a user reads docs/cli/index.md
- **THEN** an "Economy" section SHALL appear with 5 command entries after the P2P Network section

### Requirement: Contract commands in CLI reference
The docs/cli/index.md SHALL include a Contract section with a table listing all 3 contract commands: `lango contract read`, `lango contract call`, and `lango contract abi load`.

#### Scenario: Contract table exists in CLI index
- **WHEN** a user reads docs/cli/index.md
- **THEN** a "Contract" section SHALL appear with 3 command entries after the Economy section

### Requirement: Metrics commands in CLI reference
The docs/cli/index.md SHALL include a Metrics section with a table listing all 5 metrics commands: `lango metrics`, `lango metrics sessions`, `lango metrics tools`, `lango metrics agents`, and `lango metrics history`.

#### Scenario: Metrics table exists in CLI index
- **WHEN** a user reads docs/cli/index.md
- **THEN** a "Metrics" section SHALL appear with 5 command entries after the Contract section

### Requirement: CLI production code uses command-stream-safe output paths
CLI production code SHALL avoid raw process-global print calls and direct standard-stream references except where explicit seam defaults are intentionally defined.

#### Scenario: CLI production code rejects raw fmt.Print calls
- **WHEN** a non-test Go file under `internal/cli` reintroduces `fmt.Print`, `fmt.Printf`, or `fmt.Println`
- **THEN** an executable repository test SHALL fail

#### Scenario: CLI production code rejects direct standard streams outside seams
- **WHEN** a non-test Go file under `internal/cli` reintroduces direct `os.Stdout` or `os.Stderr` references outside approved seam files
- **THEN** an executable repository test SHALL fail

### Requirement: Top-level utility commands write success output through Cobra streams
Top-level utility commands under `lango` SHALL route successful human-readable output through the Cobra command output stream so wrappers and command-level tests can capture it without intercepting process-global stdout.

#### Scenario: Utility subcommands ignore the root mode flag
- **WHEN** `lango version` or `lango health` is executed with the root-level `--mode` flag
- **THEN** the utility subcommand SHALL still complete normally

### Requirement: Serve startup output writes through Cobra streams
Top-level startup commands that emit non-error banner or summary output SHALL route that success output through the Cobra command output stream.

#### Scenario: Serve startup banner and summary write to command output
- **WHEN** `lango serve` successfully starts the application
- **THEN** the startup banner and feature summary SHALL be written through the Cobra command output stream

### Requirement: Root entrypoint failure paths remain seam-aware
The top-level `lango` entrypoint SHALL route broker-mode and root-command failure messages through the configured stderr seam and produce deterministic exit codes under test. Sandbox worker mode SHALL return the worker seam's exit code without evaluating broker mode or constructing the root command.

#### Scenario: Worker mode returns worker exit code
- **WHEN** sandbox worker mode is active
- **THEN** the entrypoint SHALL invoke the sandbox worker seam
- **AND** it SHALL return the worker seam's exit code
- **AND** it SHALL NOT evaluate broker mode or construct the root command

### Requirement: TUI startup notices remain seam-aware
Interactive top-level TUI entrypoints SHALL route their startup notice text through seam-aware stderr writers so wrapper and regression captures do not depend on process-global stderr interception.

#### Scenario: Cockpit rejects non-interactive startup cleanly
- **WHEN** `lango cockpit` is executed in a non-interactive environment
- **THEN** it SHALL return an actionable error requiring an interactive terminal

#### Scenario: Chat rejects non-interactive startup cleanly
- **WHEN** `lango chat` is executed in a non-interactive environment
- **THEN** it SHALL return an actionable error requiring an interactive terminal

#### Scenario: Interactive top-level mode validation fails before app build
- **WHEN** `lango`, `lango cockpit`, or `lango chat` receives an unknown `--mode` value on an interactive path
- **THEN** it SHALL return an actionable unknown-mode error
- **AND** it SHALL reject the mode before constructing the corresponding TUI application

### Requirement: Cmd entrypoint stream routing stays disciplined
Top-level binary entrypoints SHALL avoid raw print calls and direct standard-stream references except where explicit seam declarations intentionally define default process streams.

#### Scenario: Cmd entrypoints reject raw print calls
- **WHEN** a non-test Go file under `cmd/` reintroduces `fmt.Print`, `fmt.Printf`, or `fmt.Println`
- **THEN** an executable repository test SHALL fail

#### Scenario: Cmd entrypoints reject direct standard streams outside seams
- **WHEN** a non-test Go file under `cmd/` reintroduces direct `os.Stdin`, `os.Stdout`, or `os.Stderr` references outside the approved seam declaration lines
- **THEN** an executable repository test SHALL fail

### Requirement: Cmd entrypoint exit routing stays disciplined
Top-level binary entrypoints SHALL not call `os.Exit(...)` directly except where explicit seam declarations intentionally define the default process-exit function.

#### Scenario: Cmd entrypoints reject direct os.Exit outside seams
- **WHEN** a non-test Go file under `cmd/` reintroduces a direct `os.Exit(...)` reference outside the approved seam declaration lines
- **THEN** an executable repository test SHALL fail
