## MODIFIED Requirements

### Requirement: P2P skill definitions
The skills directory SHALL include skill definitions for P2P reputation, pricing, and owner shield operations.

#### Scenario: p2p-reputation skill exists
- **WHEN** system loads skills from `skills/` directory
- **THEN** `skills/p2p-reputation/SKILL.md` exists with type `script`, status `active`, and command `lango p2p reputation --peer-did "$PEER_DID"`

#### Scenario: p2p-pricing skill exists
- **WHEN** system loads skills from `skills/` directory
- **THEN** `skills/p2p-pricing/SKILL.md` exists with type `script`, status `active`, and command `lango p2p pricing`

#### Scenario: p2p-owner-shield skill exists
- **WHEN** system loads skills from `skills/` directory
- **THEN** `skills/p2p-owner-shield/SKILL.md` exists with type `script`, status `active`, and command `lango p2p status --json | jq '.ownerShield'`
