## Purpose

P2P REST API endpoints on the gateway that expose the running P2P node's status, connected peers, and DID identity without creating ephemeral libp2p nodes.

## Requirements

### Requirement: P2P status endpoint
The gateway SHALL expose `GET /api/p2p/status` that returns the local node's peer ID, listen addresses, and connected peer count as JSON.

#### Scenario: Query P2P status when node is running
- **WHEN** a client sends `GET /api/p2p/status` to the gateway
- **THEN** the response SHALL be HTTP 200 with JSON containing `peerId` (string), `listenAddrs` (string array), and `connectedPeers` (integer)

### Requirement: P2P peers endpoint
The gateway SHALL expose `GET /api/p2p/peers` that returns a list of currently connected peers with their IDs and multiaddresses.

#### Scenario: Query connected peers
- **WHEN** a client sends `GET /api/p2p/peers` to the gateway
- **THEN** the response SHALL be HTTP 200 with JSON containing `peers` (array of objects with `peerId` and `addrs` fields) and `count` (integer)

### Requirement: P2P identity endpoint
The gateway SHALL expose `GET /api/p2p/identity` that returns the local DID string derived from the wallet.

#### Scenario: Query identity with wallet configured
- **WHEN** a client sends `GET /api/p2p/identity` and the identity provider is available
- **THEN** the response SHALL be HTTP 200 with JSON containing `did` (string starting with `did:lango:`) and `peerId` (string)

#### Scenario: Query identity without identity provider
- **WHEN** a client sends `GET /api/p2p/identity` and the identity provider is nil
- **THEN** the response SHALL be HTTP 200 with JSON containing `did` as null and `peerId` (string)

### Requirement: P2P reputation endpoint
The gateway SHALL expose `GET /api/p2p/reputation` that returns peer reputation details.

#### Scenario: GET /api/p2p/reputation with valid peer_did
- **WHEN** client sends `GET /api/p2p/reputation?peer_did=did:lango:abc123`
- **THEN** server returns JSON with full PeerDetails (peerDid, trustScore, successfulExchanges, failedExchanges, timeoutCount, firstSeen, lastInteraction)

#### Scenario: GET /api/p2p/reputation without peer_did
- **WHEN** client sends `GET /api/p2p/reputation` without peer_did query parameter
- **THEN** server returns 400 with error message "peer_did query parameter is required"

#### Scenario: GET /api/p2p/reputation for unknown peer
- **WHEN** client sends `GET /api/p2p/reputation?peer_did=did:lango:unknown`
- **THEN** server returns JSON with trustScore 0.0 and "no reputation record found" message

### Requirement: P2P pricing endpoint
The gateway SHALL expose `GET /api/p2p/pricing` that returns P2P tool pricing configuration.

#### Scenario: GET /api/p2p/pricing without tool filter
- **WHEN** client sends `GET /api/p2p/pricing`
- **THEN** server returns JSON with enabled status, perQuery default price, toolPrices map, and currency

#### Scenario: GET /api/p2p/pricing with tool filter
- **WHEN** client sends `GET /api/p2p/pricing?tool=knowledge_search`
- **THEN** server returns JSON with tool name, specific price (or default), and currency

### Requirement: P2P routes registration
The P2P REST endpoints SHALL be registered on the gateway router only when P2P components are initialized (i.e., `p2pComponents` is non-nil).

#### Scenario: P2P disabled
- **WHEN** P2P is disabled in configuration
- **THEN** no `/api/p2p/*` routes SHALL be registered on the gateway
