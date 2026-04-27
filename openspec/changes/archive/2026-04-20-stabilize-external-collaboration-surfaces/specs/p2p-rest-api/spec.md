## MODIFIED Requirements

### Requirement: P2P identity endpoint
The gateway SHALL expose `GET /api/p2p/identity` that returns the active local DID, when available, together with the libp2p peer ID.

#### Scenario: Query identity with active DID available
- **WHEN** a client sends `GET /api/p2p/identity` and the runtime can resolve an active DID
- **THEN** the response SHALL be HTTP 200 with JSON containing `did` (string starting with `did:lango:`) and `peerId` (string)

#### Scenario: Query identity without active DID
- **WHEN** a client sends `GET /api/p2p/identity` and the runtime cannot resolve an active DID
- **THEN** the response SHALL be HTTP 200 with JSON containing `did` as `null` and `peerId` (string)
