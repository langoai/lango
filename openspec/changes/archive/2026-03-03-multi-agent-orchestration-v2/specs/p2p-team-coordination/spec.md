## ADDED Requirements

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
