## Purpose

Capability spec for cli-p2p-management. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: P2P CLI command group
The system SHALL provide a `lango p2p` command group with subcommands for P2P network management, wired into `cmd/lango/main.go` using the bootstrap Result loader pattern.

#### Scenario: Root command shows help
- **WHEN** user runs `lango p2p`
- **THEN** system displays help text listing all available P2P subcommands

### Requirement: P2P status command
The system SHALL provide `lango p2p status [--output table|json]` that displays node peer ID, listen addresses, connected peer count, max peers, mDNS status, relay status, and ZK handshake status.

#### Scenario: Status in text format
- **WHEN** user runs `lango p2p status`
- **THEN** system prints peer ID, listen addrs, connected peers count, and feature flags in human-readable format

#### Scenario: Status in JSON format
- **WHEN** user runs `lango p2p status --output json`
- **THEN** system outputs a JSON object with fields: peerId, listenAddrs, connectedPeers, maxPeers, mdns, relay, zkHandshake

#### Scenario: Status rejects unknown output format before bootstrap
- **WHEN** user runs `lango p2p status --output yaml`
- **THEN** the command returns `unknown output format "yaml" (expected: table or json)`
- **AND** it does not invoke bootstrap loading

### Requirement: P2P status output routing
`lango p2p status` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture text and JSON output without intercepting process-global stdout.

#### Scenario: Status text output writes to command output
- **WHEN** user runs `lango p2p status`
- **THEN** the command writes the status text output to the Cobra command output stream

#### Scenario: Status JSON output writes to command output
- **WHEN** user runs `lango p2p status --output json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

### Requirement: P2P peers command
The system SHALL provide `lango p2p peers [--output table|json]` that lists all connected peers with peer ID and remote multiaddrs using tabwriter output.

#### Scenario: No connected peers
- **WHEN** user runs `lango p2p peers` with no connected peers
- **THEN** system prints "No connected peers."

#### Scenario: Connected peers in table format
- **WHEN** user runs `lango p2p peers` with connected peers
- **THEN** system prints a table with PEER ID and ADDRESS columns

### Requirement: P2P peers output routing
`lango p2p peers` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture empty-state, table, and JSON output without intercepting process-global stdout.

#### Scenario: No connected peers writes to command output
- **WHEN** user runs `lango p2p peers` with no connected peers
- **THEN** the command writes `No connected peers.` to the Cobra command output stream

#### Scenario: Peers table writes to command output
- **WHEN** user runs `lango p2p peers` with connected peers
- **THEN** the command writes the peers table to the Cobra command output stream

#### Scenario: Peers JSON writes to command output
- **WHEN** user runs `lango p2p peers --output json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Peers reject unknown output format before bootstrap
- **WHEN** user runs `lango p2p peers --output yaml`
- **THEN** the command returns `unknown output format "yaml" (expected: table or json)`
- **AND** it does not invoke bootstrap loading

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

### Requirement: P2P connect and disconnect output routing
`lango p2p connect` and `lango p2p disconnect` SHALL write success confirmations through the Cobra command output stream so wrappers and test harnesses can capture them without intercepting process-global stdout.

#### Scenario: Connect confirmation writes to command output
- **WHEN** user runs `lango p2p connect /ip4/1.2.3.4/tcp/9000/p2p/QmPeerId`
- **THEN** the command writes `Connected to peer QmPeerId` to the Cobra command output stream

#### Scenario: Disconnect confirmation writes to command output
- **WHEN** user runs `lango p2p disconnect QmPeerId`
- **THEN** the command writes `Disconnected from peer QmPeerId` to the Cobra command output stream

### Requirement: P2P provenance command group
The system SHALL provide `lango p2p provenance push` and `lango p2p provenance fetch` for exchanging signed provenance bundles through the running gateway using authenticated P2P sessions.

#### Scenario: Provenance push
- **WHEN** user runs `lango p2p provenance push <peer-did> <session-key>`
- **THEN** the command pushes a provenance bundle through the gateway and reports success for the target peer DID

#### Scenario: Provenance fetch
- **WHEN** user runs `lango p2p provenance fetch <peer-did> <session-key>`
- **THEN** the command fetches a provenance bundle through the gateway and reports success for the target peer DID

### Requirement: P2P provenance output routing
`lango p2p provenance push` and `lango p2p provenance fetch` SHALL write success confirmations through the Cobra command output stream so wrappers and test harnesses can capture them without intercepting process-global stdout.

#### Scenario: Provenance push confirmation writes to command output
- **WHEN** user runs `lango p2p provenance push <peer-did> <session-key>`
- **THEN** the command writes the push confirmation to the Cobra command output stream

#### Scenario: Provenance fetch confirmation writes to command output
- **WHEN** user runs `lango p2p provenance fetch <peer-did> <session-key>`
- **THEN** the command writes the fetch confirmation to the Cobra command output stream

### Requirement: P2P pricing command
The system SHALL provide `lango p2p pricing [--tool <name>] [--output table|json]` that displays provider-side P2P quote configuration including the default per-query price and tool-specific quote overrides.

#### Scenario: Pricing overview in text format
- **WHEN** user runs `lango p2p pricing`
- **THEN** the command prints whether pricing is enabled, the default per-query price, and any tool-specific prices

#### Scenario: Single tool pricing in text format
- **WHEN** user runs `lango p2p pricing --tool knowledge_search`
- **THEN** the command prints the selected tool and its public quote

#### Scenario: Pricing in JSON format
- **WHEN** user runs `lango p2p pricing --output json`
- **THEN** the JSON output SHALL include `enabled`, `perQuery`, `toolPrices`, and `currency`

#### Scenario: Pricing rejects unknown output format before bootstrap
- **WHEN** user runs `lango p2p pricing --output yaml`
- **THEN** the command returns `unknown output format "yaml" (expected: table or json)`
- **AND** it does not invoke bootstrap loading

### Requirement: P2P pricing output routing
`lango p2p pricing` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture text and JSON output without intercepting process-global stdout.

#### Scenario: Pricing text output writes to command output
- **WHEN** user runs `lango p2p pricing`
- **THEN** the command writes the pricing overview to the Cobra command output stream

#### Scenario: Pricing tool output writes to command output
- **WHEN** user runs `lango p2p pricing --tool knowledge_search`
- **THEN** the command writes the tool-specific price view to the Cobra command output stream

#### Scenario: Pricing JSON output writes to command output
- **WHEN** user runs `lango p2p pricing --output json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

### Requirement: P2P firewall command group
The system SHALL provide `lango p2p firewall [list|add|remove]` subcommands for managing knowledge firewall ACL rules.

#### Scenario: Firewall list shows config rules
- **WHEN** user runs `lango p2p firewall list`
- **THEN** system displays configured firewall rules in a table with PEER DID, ACTION, TOOLS, and RATE LIMIT columns

#### Scenario: Firewall add prints runtime-only notice
- **WHEN** user runs `lango p2p firewall add --peer-did "did:lango:02abc" --action allow`
- **THEN** system prints the rule details and a notice to persist via configuration

### Requirement: P2P firewall output routing
`lango p2p firewall` subcommands SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture empty-state, table, JSON, and guidance output without intercepting process-global stdout.

#### Scenario: Firewall list empty-state writes to command output
- **WHEN** user runs `lango p2p firewall list` with no configured rules
- **THEN** the command writes the empty-state message to the Cobra command output stream

#### Scenario: Firewall list table writes to command output
- **WHEN** user runs `lango p2p firewall list` with configured rules
- **THEN** the command writes the rules table to the Cobra command output stream

#### Scenario: Firewall list JSON writes to command output
- **WHEN** user runs `lango p2p firewall list --output json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Firewall list rejects unknown output format before bootstrap
- **WHEN** user runs `lango p2p firewall list --output yaml`
- **THEN** the command returns `unknown output format "yaml" (expected: table or json)`
- **AND** it does not invoke bootstrap loading

#### Scenario: Firewall add guidance writes to command output
- **WHEN** user runs `lango p2p firewall add ...`
- **THEN** the command writes the rule summary and persistence guidance to the Cobra command output stream

#### Scenario: Firewall remove guidance writes to command output
- **WHEN** user runs `lango p2p firewall remove <peer-did>`
- **THEN** the command writes the runtime removal guidance to the Cobra command output stream

### Requirement: P2P discover command
The system SHALL provide `lango p2p discover [--tag <tag>] [--output table|json]` that creates a GossipService and searches for agents by capability.

#### Scenario: Discover with tag filter
- **WHEN** user runs `lango p2p discover --tag research`
- **THEN** system displays agents matching the "research" capability in a table with NAME, DID, CAPABILITIES, and PEER ID columns

### Requirement: P2P discover output routing
`lango p2p discover` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture empty-state, table, and JSON output without intercepting process-global stdout.

#### Scenario: Discover empty-state writes to command output
- **WHEN** user runs `lango p2p discover` and no agents are known
- **THEN** the command writes `No agents discovered. Try connecting to bootstrap peers first.` to the Cobra command output stream

#### Scenario: Discover table writes to command output
- **WHEN** user runs `lango p2p discover --tag research`
- **THEN** the command writes the discovery table to the Cobra command output stream

#### Scenario: Discover JSON writes to command output
- **WHEN** user runs `lango p2p discover --output json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Discover rejects unknown output format before bootstrap
- **WHEN** user runs `lango p2p discover --output yaml`
- **THEN** the command returns `unknown output format "yaml" (expected: table or json)`
- **AND** it does not invoke bootstrap loading

### Requirement: P2P identity command
The system SHALL provide `lango p2p identity [--output table|json]` that displays the active DID when available, the local peer ID, key storage mode, and listen addresses.

#### Scenario: Identity in text format
- **WHEN** user runs `lango p2p identity`
- **THEN** the command SHALL print the active DID when one is available
- **AND** it SHALL print peer ID, key storage mode, and listen addresses

#### Scenario: Identity in JSON format
- **WHEN** user runs `lango p2p identity --output json`
- **THEN** the JSON output SHALL include keys `did`, `peerId`, `listenAddrs`, and `keyStorage`
- **AND** `did` SHALL be `null` when no active DID is available

#### Scenario: Identity rejects unknown output format before bootstrap
- **WHEN** user runs `lango p2p identity --output yaml`
- **THEN** the command returns `unknown output format "yaml" (expected: table or json)`
- **AND** it does not invoke bootstrap loading

### Requirement: P2P identity output routing
`lango p2p identity` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture text and JSON output without intercepting process-global stdout.

#### Scenario: Identity text output writes to command output
- **WHEN** user runs `lango p2p identity`
- **THEN** the command writes the identity text output to the Cobra command output stream

#### Scenario: Identity JSON output writes to command output
- **WHEN** user runs `lango p2p identity --output json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

### Requirement: P2P reputation command
The system SHALL provide `lango p2p reputation --peer-did <did> [--output table|json]` that displays a peer trust score, exchange history, and interaction timeline.

#### Scenario: Reputation in text format
- **WHEN** user runs `lango p2p reputation --peer-did did:lango:abc123`
- **THEN** the command prints peer DID, trust score, successes, failures, timeouts, first seen, and last interaction

#### Scenario: Reputation in JSON format
- **WHEN** user runs `lango p2p reputation --peer-did did:lango:abc123 --output json`
- **THEN** the JSON output SHALL include `peerDid`, `trustScore`, `successfulExchanges`, `failedExchanges`, `timeoutCount`, `firstSeen`, and `lastInteraction`

#### Scenario: Reputation missing record
- **WHEN** user runs `lango p2p reputation --peer-did did:lango:missing`
- **THEN** the command reports that no reputation record was found for that DID

### Requirement: P2P reputation output routing
`lango p2p reputation` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture missing-record, text, and JSON output without intercepting process-global stdout.

#### Scenario: Reputation missing-record output writes to command output
- **WHEN** user runs `lango p2p reputation --peer-did did:lango:missing`
- **THEN** the command writes the missing-record message to the Cobra command output stream

#### Scenario: Reputation text output writes to command output
- **WHEN** user runs `lango p2p reputation --peer-did did:lango:abc123`
- **THEN** the command writes the reputation text output to the Cobra command output stream

#### Scenario: Reputation JSON output writes to command output
- **WHEN** user runs `lango p2p reputation --peer-did did:lango:abc123 --output json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Reputation rejects unknown output format before bootstrap
- **WHEN** user runs `lango p2p reputation --peer-did did:lango:abc123 --output yaml`
- **THEN** the command returns `unknown output format "yaml" (expected: table or json)`
- **AND** it does not invoke bootstrap loading

### Requirement: P2P session command group
The system SHALL provide `lango p2p session` subcommands for listing active sessions and revoking one or all authenticated peer sessions.

#### Scenario: Session list in text format
- **WHEN** user runs `lango p2p session list`
- **THEN** the command prints active sessions with peer DID, created timestamp, expiry timestamp, and ZK verification state

#### Scenario: Session list in JSON format
- **WHEN** user runs `lango p2p session list --output json`
- **THEN** the JSON output SHALL contain the active session records

#### Scenario: Session revoke
- **WHEN** user runs `lango p2p session revoke --peer-did did:lango:abc123`
- **THEN** the command reports that the session was revoked

#### Scenario: Session revoke-all
- **WHEN** user runs `lango p2p session revoke-all`
- **THEN** the command reports that all sessions were revoked

### Requirement: P2P session output routing
`lango p2p session` subcommands SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture empty-state, table, JSON, and revoke confirmations without intercepting process-global stdout.

#### Scenario: Session list empty-state writes to command output
- **WHEN** user runs `lango p2p session list` with no active sessions
- **THEN** the command writes `No active sessions.` to the Cobra command output stream

#### Scenario: Session list table writes to command output
- **WHEN** user runs `lango p2p session list` with active sessions
- **THEN** the command writes the session table to the Cobra command output stream

#### Scenario: Session list JSON writes to command output
- **WHEN** user runs `lango p2p session list --output json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Session list rejects unknown output format before bootstrap
- **WHEN** user runs `lango p2p session list --output yaml`
- **THEN** the command returns `unknown output format "yaml" (expected: table or json)`
- **AND** it does not invoke bootstrap loading

#### Scenario: Session revoke confirmation writes to command output
- **WHEN** user runs `lango p2p session revoke --peer-did did:lango:abc123`
- **THEN** the command writes the revoke confirmation to the Cobra command output stream

#### Scenario: Session revoke-all confirmation writes to command output
- **WHEN** user runs `lango p2p session revoke-all`
- **THEN** the command writes the revoke-all confirmation to the Cobra command output stream

### Requirement: P2P disabled error
All P2P CLI commands SHALL return a clear error when `p2p.enabled` is false.

#### Scenario: P2P not enabled
- **WHEN** user runs any `lango p2p` subcommand with P2P disabled
- **THEN** system returns error "P2P networking is not enabled (set p2p.enabled = true)"

### Requirement: Team subcommand group addition
The existing `lango p2p` command group SHALL gain a new `team` subcommand group containing list, status, and disband subcommands for P2P team lifecycle management. The team subcommand group uses bootLoader for config access but does NOT initialize a full P2P node.

#### Scenario: P2P help includes team
- **WHEN** user runs `lango p2p --help`
- **THEN** the help output lists team alongside existing P2P subcommands (status, peers, connect, disconnect, firewall, discover, identity, session)

### Requirement: ZKP subcommand group addition
The existing `lango p2p` command group SHALL gain a new `zkp` subcommand group containing status and circuits subcommands for ZKP inspection. The zkp status subcommand uses cfgLoader; the zkp circuits subcommand requires no loader.

#### Scenario: P2P help includes zkp
- **WHEN** user runs `lango p2p --help`
- **THEN** the help output lists zkp alongside existing P2P subcommands

### Requirement: Existing P2P commands unaffected
The addition of team and zkp subcommand groups SHALL NOT change the behavior or registration of any existing P2P subcommands.

#### Scenario: Existing P2P status still works
- **WHEN** user runs `lango p2p status`
- **THEN** the command behaves identically to before the team/zkp additions

### Requirement: P2P disabled gating
The team and zkp status subcommands SHALL respect the existing P2P disabled error pattern: when `p2p.enabled` is false, the commands SHALL return the standard error "P2P networking is not enabled (set p2p.enabled = true)". The zkp circuits subcommand SHALL NOT be gated by p2p.enabled since it returns static data.

#### Scenario: Team command with P2P disabled
- **WHEN** user runs `lango p2p team list` with P2P disabled
- **THEN** system returns the standard P2P disabled error

#### Scenario: ZKP circuits with P2P disabled
- **WHEN** user runs `lango p2p zkp circuits` with P2P disabled
- **THEN** system still displays the circuit list since it is static data

### Requirement: P2P sandbox output routing
`lango p2p sandbox` subcommands SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture status, smoke-test, and cleanup output without intercepting process-global stdout.

#### Scenario: Sandbox status output writes to command output
- **WHEN** user runs `lango p2p sandbox status`
- **THEN** the command writes the sandbox status output to the Cobra command output stream

#### Scenario: Sandbox smoke-test output writes to command output
- **WHEN** user runs `lango p2p sandbox test`
- **THEN** the command writes the runtime-selection and smoke-test result output to the Cobra command output stream

#### Scenario: Sandbox cleanup output writes to command output
- **WHEN** user runs `lango p2p sandbox cleanup`
- **THEN** the command writes the cleanup success output to the Cobra command output stream

### Requirement: P2P zkp output routing
`lango p2p zkp status` and `lango p2p zkp circuits` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture text, table, and JSON output without intercepting process-global stdout.

#### Scenario: ZKP status text output writes to command output
- **WHEN** user runs `lango p2p zkp status`
- **THEN** the command writes the status text output to the Cobra command output stream

#### Scenario: ZKP status JSON output writes to command output
- **WHEN** user runs `lango p2p zkp status --output json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: ZKP circuits text output writes to command output
- **WHEN** user runs `lango p2p zkp circuits`
- **THEN** the command writes the circuits table to the Cobra command output stream

#### Scenario: ZKP circuits JSON output writes to command output
- **WHEN** user runs `lango p2p zkp circuits --output json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: ZKP commands reject unknown output format before work
- **WHEN** user runs `lango p2p zkp status --output yaml` or `lango p2p zkp circuits --output yaml`
- **THEN** the command returns `unknown output format "yaml" (expected: table or json)`
- **AND** it does not invoke config loading or circuit listing work

### Requirement: Team CLI is guidance-oriented until live team control exists
The `lango p2p team` CLI surface SHALL describe the current runtime honestly: teams are real runtime structures, but the CLI is guidance-oriented until full live control is implemented.

#### Scenario: Team list guidance
- **WHEN** user runs `lango p2p team list`
- **THEN** the command SHALL describe the current guidance-oriented or runtime-backed path instead of implying direct live team control

#### Scenario: Team guidance names the concrete team tools
- **WHEN** user runs `lango p2p team list`, `status`, or `disband`
- **THEN** the command SHALL point to the concrete `team_*` tool surface for the corresponding live action

### Requirement: P2P team output routing
`lango p2p team` guidance commands SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture text and JSON guidance output without intercepting process-global stdout.

#### Scenario: Team list text output writes to command output
- **WHEN** user runs `lango p2p team list`
- **THEN** the command writes the guidance text to the Cobra command output stream

#### Scenario: Team list JSON output writes to command output
- **WHEN** user runs `lango p2p team list --output json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Team status text output writes to command output
- **WHEN** user runs `lango p2p team status <team-id>`
- **THEN** the command writes the guidance text to the Cobra command output stream

#### Scenario: Team status JSON output writes to command output
- **WHEN** user runs `lango p2p team status <team-id> --output json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Team guidance commands reject unknown output format before bootstrap
- **WHEN** user runs `lango p2p team list --output yaml` or `lango p2p team status <team-id> --output yaml`
- **THEN** the command returns `unknown output format "yaml" (expected: table or json)`
- **AND** it does not invoke bootstrap loading

#### Scenario: Team disband text output writes to command output
- **WHEN** user runs `lango p2p team disband <team-id>`
- **THEN** the command writes the guidance text to the Cobra command output stream
- **AND** it SHALL NOT rely only on vague phrases such as "team runtime or agent tools"

#### Scenario: Team top-level help names the concrete team tools
- **WHEN** user runs `lango p2p team --help`
- **THEN** the help text SHALL mention the concrete `team_form`, `team_form_with_budget`, `team_status`, `team_list`, and `team_disband` tools
- **AND** it SHALL NOT rely on vague `agent/tool-backed` wording alone

### Requirement: Workspace CLI manages local workspace records
The `lango p2p workspace` CLI surface SHALL manage local collaborative workspace records through the same BoltDB-backed workspace manager used by the runtime while keeping distributed messaging and peer exchange delegated to the running server and workspace tools.

#### Scenario: Workspace create persists locally
- **WHEN** user runs `lango p2p workspace create <name>`
- **THEN** the command SHALL create a local workspace record
- **AND** it SHALL print the workspace ID, name, goal, status, and member count

#### Scenario: Workspace list reads local records
- **WHEN** user runs `lango p2p workspace list`
- **THEN** the command SHALL list locally persisted workspace records or print an empty state

#### Scenario: Workspace status reads one local record
- **WHEN** user runs `lango p2p workspace status <workspace-id>`
- **THEN** the command SHALL show the local workspace record including member details

#### Scenario: Workspace join and leave mutate local membership
- **WHEN** user runs `lango p2p workspace join <workspace-id>` or `leave <workspace-id>`
- **THEN** the command SHALL add or remove the local agent identity from the persisted workspace membership

#### Scenario: Workspace top-level help names the concrete workspace tools
- **WHEN** user runs `lango p2p workspace --help`
- **THEN** the help text SHALL mention the concrete `p2p_workspace_create`, `p2p_workspace_join`, `p2p_workspace_leave`, `p2p_workspace_list`, `p2p_workspace_status`, and `p2p_workspace_read` tools
- **AND** it SHALL NOT rely on vague `agent/tool-backed` wording alone

### Requirement: P2P workspace output routing
`lango p2p workspace` commands SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture text and JSON output without intercepting process-global stdout.

#### Scenario: Workspace create text output writes to command output
- **WHEN** user runs `lango p2p workspace create <name>`
- **THEN** the command writes the created workspace summary to the Cobra command output stream

#### Scenario: Workspace create JSON output writes to command output
- **WHEN** user runs `lango p2p workspace create <name> --output json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Workspace list text output writes to command output
- **WHEN** user runs `lango p2p workspace list`
- **THEN** the command writes the local workspace list to the Cobra command output stream

#### Scenario: Workspace list JSON output writes to command output
- **WHEN** user runs `lango p2p workspace list --output json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Workspace status text output writes to command output
- **WHEN** user runs `lango p2p workspace status <workspace-id>`
- **THEN** the command writes the workspace details to the Cobra command output stream

#### Scenario: Workspace status JSON output writes to command output
- **WHEN** user runs `lango p2p workspace status <workspace-id> --output json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Workspace commands reject unknown output format before bootstrap
- **WHEN** user runs `lango p2p workspace create <name> --output yaml`, `list --output yaml`, or `status <workspace-id> --output yaml`
- **THEN** the command returns `unknown output format "yaml" (expected: table or json)`
- **AND** it does not invoke bootstrap loading

#### Scenario: Workspace join/leave text output writes to command output
- **WHEN** user runs `lango p2p workspace join <workspace-id>` or `leave <workspace-id>`
- **THEN** the command writes the membership mutation result to the Cobra command output stream

### Requirement: Git CLI remains guidance-oriented until live control exists
The `lango p2p git` CLI surface SHALL describe the current runtime honestly: git bundle subsystems are real, but the CLI commands guide operators toward server-backed or tool-backed flows until fuller live control exists.

#### Scenario: Git bundle guidance
- **WHEN** user runs `lango p2p git push` or `lango p2p git fetch`
- **THEN** the command SHALL describe the current server-backed or tool-backed exchange path instead of implying a fully direct live CLI operation

#### Scenario: Git fetch guidance does not point to a nonexistent tool
- **WHEN** user runs `lango p2p git fetch`
- **THEN** the command SHALL direct the operator to the server-backed runtime exchange path
- **AND** it SHALL NOT mention a nonexistent `p2p_git_fetch` tool

#### Scenario: Git control guidance does not point to a nonexistent public API
- **WHEN** user runs `lango p2p git init`, `lango p2p git log`, `lango p2p git diff`, or `lango p2p git push`
- **THEN** the command SHALL direct the operator to the server-backed runtime and the real `p2p_git_*` tools
- **AND** it SHALL NOT imply a public workspace/git control API that does not exist

### Requirement: P2P git output routing
`lango p2p git` guidance commands SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture text and JSON guidance output without intercepting process-global stdout.

#### Scenario: Git guidance text writes to command output
- **WHEN** user runs `lango p2p git init`, `log`, `diff`, `push`, or `fetch`
- **THEN** the command writes its guidance text to the Cobra command output stream

#### Scenario: Git log JSON writes to command output
- **WHEN** user runs `lango p2p git log <workspace-id> --output json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Git log rejects unknown output format before bootstrap
- **WHEN** user runs `lango p2p git log <workspace-id> --output yaml`
- **THEN** the command returns `unknown output format "yaml" (expected: table or json)`
- **AND** it does not invoke bootstrap loading
