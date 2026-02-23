## Purpose

Per-peer DID trust scoring system that tracks exchange outcomes and integrates with the firewall to reject untrusted peers.

## Requirements

### Requirement: Trust Score Calculation
The system SHALL calculate peer trust scores based on exchange outcomes.

#### Scenario: Successful exchange
- **WHEN** a successful exchange is recorded for a peer
- **THEN** the peer's trust score increases

#### Scenario: Failed exchange
- **WHEN** a failed exchange is recorded for a peer
- **THEN** the peer's trust score decreases (failures weigh 2x)

#### Scenario: Timeout
- **WHEN** a timeout is recorded for a peer
- **THEN** the peer's trust score decreases (timeouts weigh 1.5x)

### Requirement: New Peer Handling
The system SHALL give new peers the benefit of the doubt.

#### Scenario: Unknown peer
- **WHEN** a peer has no reputation record
- **THEN** the peer is considered trusted (benefit of doubt)

### Requirement: Firewall Integration
The system SHALL integrate with the P2P firewall to reject untrusted peers.

#### Scenario: Peer below threshold
- **WHEN** a peer's trust score is above 0 but below the minimum threshold
- **THEN** the firewall rejects their requests

#### Scenario: Peer above threshold
- **WHEN** a peer's trust score meets or exceeds the minimum threshold
- **THEN** the firewall allows their requests

### Requirement: Persistence
The system SHALL persist reputation data in the database using Ent ORM.

### Requirement: Reputation data retrieval
The reputation Store SHALL provide a `GetDetails(ctx, peerDID)` method that returns full `PeerDetails` including PeerDID, TrustScore, SuccessfulExchanges, FailedExchanges, TimeoutCount, FirstSeen, and LastInteraction.

#### Scenario: Get details for known peer
- **WHEN** `GetDetails` is called with a peerDID that has a reputation record
- **THEN** system returns a `PeerDetails` struct populated from the ent PeerReputation entity

#### Scenario: Get details for unknown peer
- **WHEN** `GetDetails` is called with a peerDID that has no reputation record
- **THEN** system returns nil, nil (no error)

#### Scenario: Database error
- **WHEN** `GetDetails` is called and the database query fails
- **THEN** system returns nil and a wrapped error
