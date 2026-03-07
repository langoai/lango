## ADDED Requirements

### Requirement: P2P negotiation protocol
The system SHALL implement a P2P price negotiation protocol in `internal/economy/negotiation/` with a session-based lifecycle: Propose → Counter (repeated) → Accept/Reject.

#### Scenario: Initiator proposes terms
- **WHEN** an agent wants to negotiate price for a tool invocation
- **THEN** a NegotiationSession is created with Phase="proposed", Round=1, and the initial Terms

#### Scenario: Responder counters with different terms
- **WHEN** the responder sends a counter-offer with modified price
- **THEN** Phase changes to "countered", Round is incremented, and CurrentTerms is updated

#### Scenario: Responder accepts terms
- **WHEN** the responder accepts the current terms
- **THEN** Phase changes to "accepted" and the session is terminal

#### Scenario: Responder rejects terms
- **WHEN** the responder rejects the proposal
- **THEN** Phase changes to "rejected" and the session is terminal

### Requirement: NegotiationSession lifecycle
Each `NegotiationSession` SHALL track: ID, InitiatorDID, ResponderDID, Phase, CurrentTerms, Proposals (history), Round, MaxRounds, CreatedAt, UpdatedAt, ExpiresAt.

#### Scenario: Session terminal check
- **WHEN** Phase is "accepted", "rejected", "expired", or "cancelled"
- **THEN** `IsTerminal()` returns true

#### Scenario: Counter allowed within max rounds
- **WHEN** Phase is not terminal and Round < MaxRounds
- **THEN** `CanCounter()` returns true

#### Scenario: Counter blocked at max rounds
- **WHEN** Round >= MaxRounds
- **THEN** `CanCounter()` returns false (must accept or reject)

### Requirement: Negotiation phases
The system SHALL support six phases:
- `proposed`: initial offer sent by initiator
- `countered`: counter-offer sent by either party
- `accepted`: terms agreed upon
- `rejected`: terms explicitly rejected
- `expired`: session timeout reached
- `cancelled`: session cancelled by either party

#### Scenario: Phase transitions
- **WHEN** a negotiation progresses
- **THEN** valid transitions are: proposed→countered, proposed→accepted, proposed→rejected, countered→countered, countered→accepted, countered→rejected, any→expired, any→cancelled

### Requirement: Terms structure
Negotiated `Terms` SHALL contain: Price (*big.Int), Currency (string), ToolName (string), MaxLatency (time.Duration, optional), UseEscrow (bool), and EscrowID (string, optional).

#### Scenario: Terms include escrow decision
- **WHEN** the risk assessment recommends escrow
- **THEN** Terms.UseEscrow=true and EscrowID is set after escrow creation

### Requirement: Proposal and ProposalAction types
Each round of negotiation produces a `Proposal` with Action (propose/counter/accept/reject), SenderDID, Terms, Round, Reason (optional), and Timestamp.

#### Scenario: Counter-offer with reason
- **WHEN** a responder counters with a lower price
- **THEN** a Proposal is created with Action="counter", the new Terms, and a Reason explaining the counter

### Requirement: P2P message types
The negotiation protocol SHALL use the following P2P message types:
- `negotiate_propose`: initial price proposal
- `negotiate_respond`: counter-offer, accept, or reject

Both message types carry a `NegotiatePayload` containing SessionID and Proposal.

#### Scenario: Propose message sent
- **WHEN** an initiator starts negotiation
- **THEN** a P2P message with type "negotiate_propose" and NegotiatePayload is sent to the responder

#### Scenario: Respond message sent
- **WHEN** a responder counters or accepts
- **THEN** a P2P message with type "negotiate_respond" and NegotiatePayload is sent to the initiator

### Requirement: NegotiatePayload serialization
`NegotiatePayload` SHALL support JSON marshaling/unmarshaling via `Marshal()` and `UnmarshalNegotiatePayload(data)` functions.

#### Scenario: Round-trip serialization
- **WHEN** a NegotiatePayload is marshaled and unmarshaled
- **THEN** the deserialized payload is identical to the original

### Requirement: MaxRounds constraint
The system SHALL enforce `NegotiationConfig.MaxRounds` (default: 5). After MaxRounds counter-offers, the responder must accept or reject.

#### Scenario: Max rounds reached
- **WHEN** Round reaches MaxRounds
- **THEN** further counter-offers are rejected; only accept/reject is allowed

### Requirement: Session timeout
The system SHALL enforce `NegotiationConfig.Timeout` (default: 5m). Sessions that exceed the timeout transition to Phase="expired".

#### Scenario: Session expires
- **WHEN** the current time exceeds ExpiresAt
- **THEN** the session is marked as "expired" and cannot accept further proposals

### Requirement: Auto-negotiation strategy
When `NegotiationConfig.AutoNegotiate` is true, the system SHALL automatically generate counter-offers using a configurable discount strategy bounded by `MaxDiscount` (default: 0.2, meaning max 20% reduction from initial price).

#### Scenario: Auto-counter generated
- **WHEN** AutoNegotiate=true and a proposal is received with price above the agent's minimum
- **THEN** an automatic counter-offer is generated with a price between the offer and the minimum acceptable price

#### Scenario: Auto-accept when price is acceptable
- **WHEN** AutoNegotiate=true and the proposed price is at or below the agent's acceptable threshold
- **THEN** the proposal is automatically accepted

#### Scenario: Auto-reject when discount exceeds max
- **WHEN** the proposed discount exceeds MaxDiscount from the base price
- **THEN** the proposal is automatically rejected

### Requirement: NegotiationConfig defaults
The system SHALL use the following defaults from `config.NegotiationConfig`:
- `Enabled`: false (opt-in)
- `MaxRounds`: 5
- `Timeout`: 5m
- `AutoNegotiate`: false
- `MaxDiscount`: 0.2 (20%)

#### Scenario: Negotiation disabled by default
- **WHEN** NegotiationConfig.Enabled is not set
- **THEN** the system uses fixed prices without negotiation

### Requirement: Negotiation events
The system SHALL publish the following events via the event bus:
- `NegotiationStartedEvent`: when a new session is created (contains SessionID, InitiatorDID, ResponderDID, initial Terms)
- `NegotiationCompletedEvent`: when a session reaches "accepted" (contains SessionID, agreed Terms)
- `NegotiationFailedEvent`: when a session reaches "rejected", "expired", or "cancelled" (contains SessionID, Phase, Reason)

#### Scenario: Successful negotiation event
- **WHEN** a negotiation session reaches "accepted"
- **THEN** a NegotiationCompletedEvent is published with the final agreed Terms

#### Scenario: Failed negotiation event
- **WHEN** a negotiation session times out
- **THEN** a NegotiationFailedEvent is published with Phase="expired"
