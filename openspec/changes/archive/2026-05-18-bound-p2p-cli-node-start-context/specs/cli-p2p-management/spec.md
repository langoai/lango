## ADDED Requirements

### Requirement: P2P CLI ephemeral node startup uses command context
P2P CLI commands that start an ephemeral P2P node SHALL derive that node startup from the Cobra command context.

#### Scenario: Inspection commands pass command context to node startup
- **WHEN** an operator runs `lango p2p status`, `lango p2p peers`, `lango p2p discover`, or `lango p2p identity`
- **THEN** the command SHALL pass `cmd.Context()` into the shared ephemeral P2P dependency initialization path

#### Scenario: Session commands pass command context to node startup
- **WHEN** an operator runs `lango p2p session list`, `lango p2p session revoke <peer-did>`, or `lango p2p session revoke-all`
- **THEN** the command SHALL pass `cmd.Context()` into the shared ephemeral P2P dependency initialization path

#### Scenario: Connect and disconnect pass command context to node startup
- **WHEN** an operator runs `lango p2p connect <multiaddr>` or `lango p2p disconnect <peer-id>`
- **THEN** the command SHALL pass `cmd.Context()` into the shared ephemeral P2P dependency initialization path before attempting the peer-specific operation

#### Scenario: Ephemeral node cleanup waits for startup workers
- **WHEN** a P2P CLI command finishes and cleans up its ephemeral node
- **THEN** cleanup SHALL stop the node and wait for startup worker goroutines registered with the startup wait group
