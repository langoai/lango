## ADDED Requirements

### Requirement: Team coordination documentation in p2p-network.md
The system SHALL expand team coordination documentation in `docs/features/p2p-network.md` with conflict resolution strategies, assignment strategies, payment coordination, and team events.

#### Scenario: Conflict resolution strategies documented
- **WHEN** a user reads the team coordination section in p2p-network.md
- **THEN** they find descriptions of trust_weighted, majority_vote, leader_decides, and fail_on_conflict strategies

#### Scenario: Assignment strategies documented
- **WHEN** a user reads the team coordination section
- **THEN** they find descriptions of best_match, round_robin, and load_balanced assignment strategies

#### Scenario: Payment coordination documented
- **WHEN** a user reads the team coordination section
- **THEN** they find PaymentCoordinator with trust-based mode selection (free/prepay/postpay)

#### Scenario: Team events documented
- **WHEN** a user reads the team coordination section
- **THEN** they find a table of team events from `internal/eventbus/team_events.go`

### Requirement: Team CLI documentation in p2p.md
The system SHALL document team coordination features (conflict resolution, assignment, payment modes) in `docs/cli/p2p.md`.

#### Scenario: Team features in CLI docs
- **WHEN** a user reads `docs/cli/p2p.md`
- **THEN** they find notes about conflict resolution strategies, assignment strategies, and payment coordination

### Requirement: README reflects team enhancements
The system SHALL mention P2P Teams with conflict resolution in `README.md`.

#### Scenario: Team features in README
- **WHEN** a user reads README.md
- **THEN** P2P Teams with conflict resolution strategies and payment coordination are mentioned
