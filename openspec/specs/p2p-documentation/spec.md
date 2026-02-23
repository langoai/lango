## ADDED Requirements

### Requirement: P2P feature documentation
The system SHALL provide docs/features/p2p-network.md covering: overview, identity (DID scheme), handshake flow, knowledge firewall (ACL rules, response sanitization, ZK attestation), discovery (GossipSub, agent card structure), ZK circuits, configuration, and CLI commands.

#### Scenario: Feature doc exists with all sections
- **WHEN** the P2P feature documentation is opened
- **THEN** it contains sections for Overview, Identity, Handshake, Knowledge Firewall, Discovery, ZK Circuits, Configuration, and CLI Commands

### Requirement: P2P CLI reference documentation
The system SHALL provide docs/cli/p2p.md with usage, flags, arguments, and examples for all P2P commands: status, peers, connect, disconnect, firewall (list/add/remove), discover, and identity.

#### Scenario: CLI doc covers all commands
- **WHEN** the P2P CLI reference is opened
- **THEN** each P2P subcommand has its own section with usage syntax, flag table, and example output

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
