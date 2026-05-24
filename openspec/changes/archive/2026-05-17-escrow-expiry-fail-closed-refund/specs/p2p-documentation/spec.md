## MODIFIED Requirements

### Requirement: Economy feature documentation
The economy feature documentation SHALL include Hub V2, milestone settler, and dangling detector sections.

#### Scenario: Hub V2 documentation
- **WHEN** `docs/features/economy.md` is rendered
- **THEN** a Hub V2 section SHALL describe V2 deal ID mapping and V2 event handling

#### Scenario: Milestone settler documentation
- **WHEN** `docs/features/economy.md` is rendered
- **THEN** a Milestone Settler section SHALL describe milestone-aware locking and release behavior

#### Scenario: Dangling detector documentation
- **WHEN** `docs/features/economy.md` is rendered
- **THEN** a Dangling Escrow Detector section SHALL describe `DanglingDetector`, `EscrowDanglingEvent`, and ExpiresAt-gated auto-expiry
