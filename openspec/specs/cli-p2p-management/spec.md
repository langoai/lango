## ADDED Requirements

### Requirement: P2P CLI command group
The system SHALL provide a `lango p2p` command group with subcommands for P2P network management, wired into `cmd/lango/main.go` using the bootstrap Result loader pattern.

#### Scenario: Root command shows help
- **WHEN** user runs `lango p2p`
- **THEN** system displays help text listing all available P2P subcommands

### Requirement: P2P status command
The system SHALL provide `lango p2p status [--json]` that displays node peer ID, listen addresses, connected peer count, max peers, mDNS status, relay status, and ZK handshake status.

#### Scenario: Status in text format
- **WHEN** user runs `lango p2p status`
- **THEN** system prints peer ID, listen addrs, connected peers count, and feature flags in human-readable format

#### Scenario: Status in JSON format
- **WHEN** user runs `lango p2p status --json`
- **THEN** system outputs a JSON object with fields: peerId, listenAddrs, connectedPeers, maxPeers, mdns, relay, zkHandshake

### Requirement: P2P peers command
The system SHALL provide `lango p2p peers [--json]` that lists all connected peers with peer ID and remote multiaddrs using tabwriter output.

#### Scenario: No connected peers
- **WHEN** user runs `lango p2p peers` with no connected peers
- **THEN** system prints "No connected peers."

#### Scenario: Connected peers in table format
- **WHEN** user runs `lango p2p peers` with connected peers
- **THEN** system prints a table with PEER ID and ADDRESS columns

### Requirement: P2P connect command
The system SHALL provide `lango p2p connect <multiaddr>` that parses the multiaddr, extracts peer info, and connects to the peer via the libp2p host.

#### Scenario: Successful connection
- **WHEN** user runs `lango p2p connect /ip4/1.2.3.4/tcp/9000/p2p/QmPeerId`
- **THEN** system connects and prints "Connected to peer QmPeerId"

#### Scenario: Invalid multiaddr
- **WHEN** user runs `lango p2p connect invalid-addr`
- **THEN** system returns an error "parse multiaddr: ..."

### Requirement: P2P disconnect command
The system SHALL provide `lango p2p disconnect <peer-id>` that closes the connection to the specified peer.

#### Scenario: Successful disconnection
- **WHEN** user runs `lango p2p disconnect QmPeerId`
- **THEN** system closes the peer connection and prints "Disconnected from peer QmPeerId"

### Requirement: P2P firewall command group
The system SHALL provide `lango p2p firewall [list|add|remove]` subcommands for managing knowledge firewall ACL rules.

#### Scenario: Firewall list shows config rules
- **WHEN** user runs `lango p2p firewall list`
- **THEN** system displays configured firewall rules in a table with PEER DID, ACTION, TOOLS, and RATE LIMIT columns

#### Scenario: Firewall add prints runtime-only notice
- **WHEN** user runs `lango p2p firewall add --peer-did "did:lango:02abc" --action allow`
- **THEN** system prints the rule details and a notice to persist via configuration

### Requirement: P2P discover command
The system SHALL provide `lango p2p discover [--tag <tag>] [--json]` that creates a GossipService and searches for agents by capability.

#### Scenario: Discover with tag filter
- **WHEN** user runs `lango p2p discover --tag research`
- **THEN** system displays agents matching the "research" capability in a table with NAME, DID, CAPABILITIES, and PEER ID columns

### Requirement: P2P identity command
The system SHALL provide `lango p2p identity [--json]` that displays the local peer ID, key directory, and listen addresses.

#### Scenario: Identity in text format
- **WHEN** user runs `lango p2p identity`
- **THEN** system prints peer ID, key directory path, and listen addresses

### Requirement: P2P disabled error
All P2P CLI commands SHALL return a clear error when `p2p.enabled` is false.

#### Scenario: P2P not enabled
- **WHEN** user runs any `lango p2p` subcommand with P2P disabled
- **THEN** system returns error "P2P networking is not enabled (set p2p.enabled = true)"
