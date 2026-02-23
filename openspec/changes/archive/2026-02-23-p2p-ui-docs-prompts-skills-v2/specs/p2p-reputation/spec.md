## MODIFIED Requirements

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
