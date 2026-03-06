## ADDED Requirements

### Requirement: Negotiation session lifecycle
The system SHALL support propose → counter → accept/reject negotiation flow. Sessions SHALL track round count, max rounds, expiration, and current terms (toolName, price, currency).

#### Scenario: Propose creates session
- **WHEN** Propose is called with initiatorDID, responderDID, and terms
- **THEN** a NegotiationSession is created with Phase=proposed, Round=1

#### Scenario: Counter increments round
- **WHEN** Counter is called on a proposed session
- **THEN** Phase transitions to "countered" and Round increments

#### Scenario: Accept finalizes
- **WHEN** Accept is called on a proposed or countered session
- **THEN** Phase transitions to "accepted" (terminal)

#### Scenario: Reject terminates
- **WHEN** Reject is called with a reason
- **THEN** Phase transitions to "rejected" (terminal)

### Requirement: Turn-based validation
The system SHALL enforce alternating turns — the same sender MUST NOT act twice in a row. The system SHALL validate that the sender is a participant (initiator or responder).

#### Scenario: Same sender acts twice
- **WHEN** the last proposal sender tries to counter again
- **THEN** ErrNotYourTurn is returned

#### Scenario: Non-participant acts
- **WHEN** a DID not matching initiator or responder tries to act
- **THEN** ErrInvalidSender is returned

### Requirement: Session expiry
The system SHALL expire sessions that exceed the configured timeout. CheckExpiry SHALL transition expired sessions and return their IDs.

#### Scenario: Session expires
- **WHEN** CheckExpiry is called and a session has passed its ExpiresAt
- **THEN** the session Phase transitions to "expired"

### Requirement: Auto-negotiation
The system SHALL support AutoRespond that uses pricing and maxDiscount to automatically accept, counter, or reject proposals.

#### Scenario: Auto-accept at base price
- **WHEN** AutoRespond is called and proposed price >= base price
- **THEN** the session is accepted

#### Scenario: Auto-counter below floor
- **WHEN** proposed price < minPrice but rounds remain
- **THEN** a counter-offer is generated using midpoint strategy

### Requirement: P2P protocol integration
The system SHALL handle RequestNegotiatePropose and RequestNegotiateRespond message types through the P2P protocol handler via a NegotiateHandler callback.

#### Scenario: Remote propose via P2P
- **WHEN** a RequestNegotiatePropose message arrives with action="propose"
- **THEN** a new negotiation session is created and session ID is returned
