## ADDED Requirements

### Requirement: p2p_price_query agent tool
The system SHALL provide a `p2p_price_query` agent tool with SafetyLevel Safe that queries remote peer pricing.

#### Scenario: Query price for a tool
- **WHEN** agent invokes `p2p_price_query` with `peer_did` and `tool_name`
- **THEN** system looks up active session, creates RemoteAgent, calls QueryPrice, and returns PriceQuoteResult with toolName, price, currency, isFree

#### Scenario: No active session
- **WHEN** agent invokes `p2p_price_query` with a peer_did that has no active session
- **THEN** system returns error "no active session for peer — connect first"

### Requirement: p2p_reputation agent tool
The system SHALL provide a `p2p_reputation` agent tool with SafetyLevel Safe that checks peer trust scores.

#### Scenario: Check reputation for known peer
- **WHEN** agent invokes `p2p_reputation` with `peer_did` for a peer with reputation data
- **THEN** system returns trustScore, isTrusted, successfulExchanges, failedExchanges, timeoutCount, firstSeen, lastInteraction

#### Scenario: Check reputation for new peer
- **WHEN** agent invokes `p2p_reputation` with `peer_did` for a peer with no reputation record
- **THEN** system returns score 0.0, isTrusted true, and "new peer" message

#### Scenario: Reputation system unavailable
- **WHEN** agent invokes `p2p_reputation` but reputation store is nil (no database)
- **THEN** system returns error "reputation system not available"
