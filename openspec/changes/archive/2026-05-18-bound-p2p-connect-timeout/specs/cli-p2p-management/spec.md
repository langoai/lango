## ADDED Requirements

### Requirement: P2P connect uses bounded command context
`lango p2p connect <multiaddr>` SHALL connect with a bounded context derived from the Cobra command context.

#### Scenario: Connect passes command cancellation to host connect
- **WHEN** the command context is canceled while `lango p2p connect <multiaddr>` is connecting
- **THEN** the host connect attempt SHALL observe the cancellation
- **AND** the command SHALL return a canceled error for the target peer

#### Scenario: Connect reports parent deadline separately
- **WHEN** the command context deadline expires before the configured connect timeout
- **THEN** the command SHALL return an error that identifies the command context deadline as the cause
- **AND** the error SHALL NOT report the configured timeout as the elapsed timeout

#### Scenario: Connect timeout uses P2P handshake timeout
- **WHEN** `p2p.handshakeTimeout` is configured to a positive duration
- **THEN** `lango p2p connect <multiaddr>` SHALL use that duration as the connect timeout
- **AND** timeout errors SHALL include the peer ID and configured timeout duration

#### Scenario: Configured timeout remains distinct from later parent deadline
- **WHEN** the configured connect timeout expires before a parent command deadline
- **THEN** the command SHALL report the configured timeout duration as the cause
- **AND** the error SHALL NOT report the parent command deadline as the cause

#### Scenario: Connect timeout falls back when unset
- **WHEN** `p2p.handshakeTimeout` is zero or negative
- **THEN** `lango p2p connect <multiaddr>` SHALL use a 30 second connect timeout

#### Scenario: Connect failure cleans up ephemeral node
- **WHEN** the connect attempt returns timeout, cancellation, or any other error
- **THEN** the ephemeral P2P node cleanup SHALL still run
