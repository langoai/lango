## MODIFIED Requirements

### Requirement: Example configs include paid value exchange settings
The P2P trading example configs SHALL include pricing, owner protection, and minimum trust score settings.

#### Scenario: Alice config has pricing enabled
- **WHEN** user reads `examples/p2p-trading/configs/alice.json`
- **THEN** P2P section includes `pricing` object with enabled=true, perQuery="0.10", and toolPrices map

#### Scenario: Alice config has owner protection
- **WHEN** user reads `examples/p2p-trading/configs/alice.json`
- **THEN** P2P section includes `ownerProtection` object with ownerName="Alice" and blockConversations=true

#### Scenario: All configs have minTrustScore
- **WHEN** user reads any of alice.json, bob.json, or charlie.json
- **THEN** P2P section includes `minTrustScore: 0.3`

#### Scenario: Each agent has correct ownerName
- **WHEN** user reads bob.json
- **THEN** ownerProtection.ownerName is "Bob"
- **WHEN** user reads charlie.json
- **THEN** ownerProtection.ownerName is "Charlie"
