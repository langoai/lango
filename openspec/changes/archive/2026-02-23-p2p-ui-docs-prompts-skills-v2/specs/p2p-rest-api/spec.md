## MODIFIED Requirements

### Requirement: P2P REST API endpoints
The P2P REST API SHALL expose reputation and pricing endpoints alongside existing status, peers, and identity endpoints.

#### Scenario: GET /api/p2p/reputation with valid peer_did
- **WHEN** client sends `GET /api/p2p/reputation?peer_did=did:lango:abc123`
- **THEN** server returns JSON with full PeerDetails (peerDid, trustScore, successfulExchanges, failedExchanges, timeoutCount, firstSeen, lastInteraction)

#### Scenario: GET /api/p2p/reputation without peer_did
- **WHEN** client sends `GET /api/p2p/reputation` without peer_did query parameter
- **THEN** server returns 400 with error message "peer_did query parameter is required"

#### Scenario: GET /api/p2p/reputation for unknown peer
- **WHEN** client sends `GET /api/p2p/reputation?peer_did=did:lango:unknown`
- **THEN** server returns JSON with trustScore 0.0 and "no reputation record found" message

#### Scenario: GET /api/p2p/pricing without tool filter
- **WHEN** client sends `GET /api/p2p/pricing`
- **THEN** server returns JSON with enabled status, perQuery default price, toolPrices map, and currency

#### Scenario: GET /api/p2p/pricing with tool filter
- **WHEN** client sends `GET /api/p2p/pricing?tool=knowledge_search`
- **THEN** server returns JSON with tool name, specific price (or default), and currency
