## MODIFIED Requirements

### Requirement: P2P provenance command group
The system SHALL provide `lango p2p provenance push` and `lango p2p provenance fetch` for exchanging signed provenance bundles through the running gateway using authenticated P2P sessions. When `--addr` is supplied, the command SHALL use the normalized explicit gateway address. When `--addr` is omitted, the command SHALL use the gateway address resolved from configured `server.host` and `server.port`, including bracket-safe IPv6 host formatting.

#### Scenario: P2P provenance explicit address is normalized
- **WHEN** user runs `lango p2p provenance push <peer-did> <session-key> --addr http://127.0.0.1:18789/`
- **THEN** the command SHALL post to the gateway using `http://127.0.0.1:18789`

#### Scenario: P2P provenance uses configured gateway when address omitted
- **WHEN** user runs `lango p2p provenance fetch <peer-did> <session-key>` with configured `server.host` and `server.port`
- **THEN** the command SHALL post to the configured gateway address
