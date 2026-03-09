## Purpose

Distributed agent team coordination for P2P network. Manages team lifecycle (forming, delegation, result collection, disbanding), conflict resolution strategies, and team events.
## Requirements
### Requirement: Team and Member types
The `p2p/team` package SHALL define `Team`, `Member`, `TeamState`, `MemberRole`, and `MemberStatus` types for representing distributed agent teams.

#### Scenario: Team lifecycle states
- **WHEN** a Team is created
- **THEN** it SHALL progress through states: Forming → Active → Completing → Disbanded

#### Scenario: Member roles
- **WHEN** members join a team
- **THEN** each SHALL have a role: Leader, Worker, or Observer

### Requirement: TeamCoordinator
The `Coordinator` SHALL provide methods: FormTeam, DelegateTask, CollectResults, DisbandTeam. It SHALL manage the full team lifecycle.

#### Scenario: Form team
- **WHEN** FormTeam is called with a list of member DIDs
- **THEN** a new Team SHALL be created and members SHALL be assigned roles

#### Scenario: Delegate task
- **WHEN** DelegateTask is called with a task description and team ID
- **THEN** the task SHALL be assigned to the best-scoring member via the Selector

#### Scenario: Collect results
- **WHEN** CollectResults is called after task delegation
- **THEN** it SHALL return results from all members that completed their tasks

#### Scenario: Disband team
- **WHEN** DisbandTeam is called
- **THEN** the team state SHALL transition to Disbanded and all members SHALL be released

### Requirement: Conflict resolution strategies
The Coordinator SHALL support multiple conflict resolution strategies: TrustWeighted (default), MajorityVote, LeaderDecides, FailOnConflict.

#### Scenario: TrustWeighted resolution
- **WHEN** multiple members return conflicting results with TrustWeighted strategy
- **THEN** the result from the member with the highest trust score SHALL be selected

#### Scenario: MajorityVote resolution
- **WHEN** multiple members return results with MajorityVote strategy
- **THEN** the most common result SHALL be selected

### Requirement: Team events
The Coordinator SHALL publish events via EventBus: TeamMemberJoinedEvent, TeamMemberLeftEvent when members join or leave a team.

#### Scenario: Member joined event
- **WHEN** a member joins a team
- **THEN** a TeamMemberJoinedEvent SHALL be published with TeamID and MemberDID

#### Scenario: Member left event
- **WHEN** a member leaves a team
- **THEN** a TeamMemberLeftEvent SHALL be published with TeamID, MemberDID, and Reason

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

