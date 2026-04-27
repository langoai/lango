## Purpose

Capability spec for p2p-documentation. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: P2P feature documentation
The system SHALL provide docs/features/p2p-network.md covering: overview, identity (DID scheme), handshake flow, knowledge firewall (ACL rules, response sanitization, ZK attestation), discovery (GossipSub, agent card structure), ZK circuits, configuration, and CLI commands. The documentation SHALL describe the live external-collaboration behavior truthfully, including current DID identity modes, payment-surface ownership, and guidance-oriented team/workspace/git operator surfaces where direct live control does not yet exist.

#### Scenario: Feature doc exists with all sections
- **WHEN** the P2P feature documentation is opened
- **THEN** it contains sections for Overview, Identity, Handshake, Knowledge Firewall, Discovery, ZK Circuits, Configuration, and CLI Commands

#### Scenario: Identity documentation includes both DID modes
- **WHEN** a user reads `docs/features/p2p-network.md`
- **THEN** the identity section SHALL describe both legacy wallet-derived `did:lango:<hex>` identities and bundle-backed `did:lango:v2:<hash>` identities

#### Scenario: Team and workspace command summary is truthful
- **WHEN** a user reads the quick command list in `docs/features/p2p-network.md`
- **THEN** team, workspace, and git commands SHALL be described as guidance-oriented or inspection-oriented when they do not provide full direct live control

#### Scenario: Workspace chronicler wording reflects partial wiring
- **WHEN** a user reads the workspace chronicler section
- **THEN** the documentation SHALL explain that graph-triple persistence depends on triple-adder wiring being available and is not yet guaranteed as a default live path

### Requirement: P2P CLI reference documentation
The system SHALL provide docs/cli/p2p.md with usage, flags, arguments, and examples for all P2P commands: status, peers, connect, disconnect, firewall (list/add/remove), discover, and identity. Examples and descriptions SHALL reflect the current runtime honestly.

#### Scenario: CLI doc covers all commands
- **WHEN** the P2P CLI reference is opened
- **THEN** each P2P subcommand has its own section with usage syntax, flag table, and example output

#### Scenario: Identity docs describe active DID exposure
- **WHEN** a user reads the `lango p2p identity` section in `docs/cli/p2p.md`
- **THEN** the documentation SHALL explain that the command exposes the active DID when one is available and SHALL distinguish legacy and v2 DID modes

#### Scenario: Pricing docs describe provider-side quote surface
- **WHEN** a user reads the `lango p2p pricing` section
- **THEN** the documentation SHALL describe it as the provider-side public quote configuration surface rather than a generic pricing policy engine

#### Scenario: Team, workspace, and git examples are guidance-oriented
- **WHEN** a user reads the `team`, `workspace`, or `git` CLI sections
- **THEN** the examples SHALL match the current server-backed or tool-backed reality instead of implying direct live control

### Requirement: README P2P sections
The README.md SHALL include P2P in the features list, CLI commands section, configuration reference table, and architecture tree.

#### Scenario: README features include P2P
- **WHEN** the README is opened
- **THEN** the Features section includes a P2P Network bullet point

#### Scenario: README CLI includes P2P commands
- **WHEN** the README CLI commands section is read
- **THEN** it lists all 9 P2P CLI commands (status, peers, connect, disconnect, firewall list/add/remove, discover, identity)

### Requirement: Features index P2P card
The docs/features/index.md SHALL include a P2P Network card in the grid layout with experimental badge and a row in the Feature Status table.

#### Scenario: Feature index includes P2P card
- **WHEN** the features index page is rendered
- **THEN** a P2P Network card appears with experimental badge linking to p2p-network.md

### Requirement: A2A protocol HTTP vs P2P comparison
The docs/features/a2a-protocol.md SHALL include a comparison section distinguishing A2A-over-HTTP from A2A-over-P2P across transport, discovery, identity, auth, firewall, and use case dimensions.

#### Scenario: A2A doc includes comparison table
- **WHEN** the A2A protocol documentation is opened
- **THEN** it contains an "A2A-over-HTTP vs A2A-over-P2P" section with a comparison table

### Requirement: P2P feature documentation includes paid value exchange
The P2P documentation SHALL include sections for Paid Value Exchange, Reputation System, and Owner Shield.

#### Scenario: p2p-network.md has Paid Value Exchange section
- **WHEN** user reads `docs/features/p2p-network.md`
- **THEN** document includes Payment Gate flow, USDC Registry description, and pricing config example

#### Scenario: p2p-network.md has Reputation System section
- **WHEN** user reads `docs/features/p2p-network.md`
- **THEN** document includes trust score formula, exchange tracking description, and querying methods (CLI/tool/API)

#### Scenario: p2p-network.md has Owner Shield section
- **WHEN** user reads `docs/features/p2p-network.md`
- **THEN** document includes PII protection description and config example

#### Scenario: configuration.md has pricing and protection config
- **WHEN** user reads `docs/configuration.md`
- **THEN** P2P section includes 9 new config fields for pricing, ownerProtection, and minTrustScore

#### Scenario: cli/p2p.md has new command references
- **WHEN** user reads `docs/cli/p2p.md`
- **THEN** document includes `reputation` and `pricing` command references with flags and examples

### Requirement: P2P network documentation covers team coordination
The P2P network feature documentation SHALL include health monitoring, graceful shutdown, git state divergence detection, reorg protection, and event-driven bridges sections.

#### Scenario: Health monitoring documented
- **WHEN** a user reads `docs/features/p2p-network.md`
- **THEN** a Health Monitoring section SHALL describe periodic pings, `maxMissed` threshold, `TeamMemberUnhealthyEvent`, and config fields

#### Scenario: Graceful shutdown documented
- **WHEN** a user reads `docs/features/p2p-network.md`
- **THEN** a Graceful Shutdown section SHALL describe `TeamGracefulShutdownEvent`, shutdown sequence, and budget settlement

#### Scenario: Reorg protection documented
- **WHEN** a user reads `docs/features/p2p-network.md`
- **THEN** a Reorg Protection section SHALL describe `confirmationDepth`, `blockHashes` cache, and `EscrowReorgDetectedEvent`

#### Scenario: Event-driven bridges documented
- **WHEN** a user reads `docs/features/p2p-network.md`
- **THEN** an Event-Driven Bridges section SHALL list all 6 bridge components

### Requirement: P2P CLI documents git bundle and task branches
The P2P CLI documentation SHALL include incremental git bundle operations and task branch management commands.

#### Scenario: Incremental bundles documented
- **WHEN** a user reads `docs/cli/p2p.md`
- **THEN** `CreateIncrementalBundle`, `VerifyBundle`, `SafeApplyBundle`, `HasCommit` SHALL be documented

#### Scenario: Task branch commands documented
- **WHEN** a user reads `docs/cli/p2p.md`
- **THEN** `p2p git branch create/list/merge/delete` commands SHALL be documented with examples

### Requirement: Economy documentation covers Escrow Hub V2
The economy feature documentation SHALL include Hub V2, milestone settler, and dangling detector sections.

#### Scenario: Hub V2 documented
- **WHEN** a user reads `docs/features/economy.md`
- **THEN** a Hub V2 section SHALL describe `HubV2Client`, `DirectSettle`, milestone/team escrows, and UUPS upgradeability

#### Scenario: Dangling detector documented
- **WHEN** a user reads `docs/features/economy.md`
- **THEN** a Dangling Escrow Detector section SHALL describe `DanglingDetector`, `EscrowDanglingEvent`, and auto-refund

### Requirement: Cron documentation covers per-job timeout
The cron automation documentation SHALL include per-job timeout configuration.

#### Scenario: Per-job timeout documented
- **WHEN** a user reads `docs/automation/cron.md`
- **THEN** a Per-Job Timeout section SHALL describe `--timeout` flag, `Timeout` field, and global vs per-job precedence
