## ADDED Requirements

### Requirement: BoltDB Persistent Team Store
The `p2p/team` package SHALL provide a `TeamStore` interface and a `BoltStore` implementation backed by BoltDB for persisting team state across process restarts.

#### Scenario: TeamStore interface
- **WHEN** the `TeamStore` interface is defined
- **THEN** it SHALL expose `Save(team)`, `Load(teamID)`, `LoadAll()`, and `Delete(teamID)` methods

#### Scenario: BoltStore creation
- **WHEN** `NewBoltStore` is called with a BoltDB instance
- **THEN** it SHALL create a "teams" bucket if it does not exist and return a ready store

#### Scenario: Save and load a team
- **WHEN** a team is saved via `BoltStore.Save()` and then loaded via `BoltStore.Load()`
- **THEN** the loaded team SHALL have identical ID, name, goal, status, budget, and members

#### Scenario: Load non-existent team
- **WHEN** `BoltStore.Load()` is called with an unknown team ID
- **THEN** it SHALL return `ErrTeamNotFound`

#### Scenario: LoadAll teams
- **WHEN** multiple teams are saved and `BoltStore.LoadAll()` is called
- **THEN** it SHALL return all persisted teams, skipping any corrupt entries with a warning

#### Scenario: Delete a team
- **WHEN** `BoltStore.Delete()` is called with a team ID
- **THEN** the team SHALL be removed from BoltDB and subsequent Load calls SHALL return `ErrTeamNotFound`

### Requirement: Coordinator persistence integration
The `Coordinator` SHALL accept a `TeamStore` in its config and auto-persist team state on lifecycle transitions.

#### Scenario: Persist on formation
- **WHEN** `FormTeam()` completes successfully and a store is configured
- **THEN** the team SHALL be persisted to the store

#### Scenario: Persist on task completion
- **WHEN** `DelegateTask()` completes and a store is configured
- **THEN** the updated team state (including spend changes) SHALL be persisted

#### Scenario: Delete on disband
- **WHEN** `DisbandTeam()` is called and a store is configured
- **THEN** the team SHALL be removed from the persistent store

#### Scenario: Load persisted teams on startup
- **WHEN** `LoadPersistedTeams()` is called during startup
- **THEN** all teams with Active or Forming status SHALL be restored into the coordinator's in-memory map

### Requirement: Coordinator membership management
The `Coordinator` SHALL support kicking members and querying team membership.

#### Scenario: Kick member
- **WHEN** `KickMember()` is called with a teamID, memberDID, and reason
- **THEN** the member SHALL be removed from the team, the state SHALL be persisted, and a `TeamMemberLeftEvent` SHALL be published

#### Scenario: Kick non-existent member
- **WHEN** `KickMember()` is called with a DID that is not in the team
- **THEN** it SHALL return `ErrNotMember`

#### Scenario: Query teams for member
- **WHEN** `TeamsForMember()` is called with a DID
- **THEN** it SHALL return all active team IDs containing that member

### Requirement: Graceful team shutdown
The `Coordinator` SHALL support graceful shutdown that blocks new tasks, settles proportional milestones, and publishes a shutdown event before disbanding.

#### Scenario: Graceful shutdown lifecycle
- **WHEN** `GracefulShutdown()` is called with a teamID and reason
- **THEN** the team status SHALL transition to `StatusShuttingDown`, a `TeamGracefulShutdownEvent` SHALL be published, and the team SHALL be disbanded

#### Scenario: Block new tasks during shutdown
- **WHEN** a team is in `StatusShuttingDown` and `DelegateTask()` is called
- **THEN** it SHALL return `ErrTeamShuttingDown`

#### Scenario: Double shutdown rejected
- **WHEN** `GracefulShutdown()` is called on a team already in `StatusShuttingDown` or `StatusDisbanded`
- **THEN** it SHALL return an error indicating the team is already in that state

### Requirement: Enhanced team model
The `Team` type SHALL support budget tracking, member enumeration, and JSON serialization of members.

#### Scenario: Budget tracking
- **WHEN** `AddSpend()` is called and the total spend exceeds the team budget
- **THEN** it SHALL return `ErrBudgetExceeded`

#### Scenario: Members enumeration
- **WHEN** `Members()` is called
- **THEN** it SHALL return deep copies of all current members safe for concurrent use

#### Scenario: MemberCount
- **WHEN** `MemberCount()` is called
- **THEN** it SHALL return the number of members in the team

#### Scenario: ActiveMembers filtering
- **WHEN** `ActiveMembers()` is called
- **THEN** it SHALL return only members not in `MemberLeft` or `MemberFailed` status

#### Scenario: JSON round-trip
- **WHEN** a team with members is marshaled to JSON and unmarshaled back
- **THEN** the members map SHALL be correctly reconstructed from the serialized members slice

### Requirement: Team shutdown and health event types
The `eventbus` package SHALL define events for graceful shutdown, health checks, member unhealthiness, budget warnings, and leader changes.

#### Scenario: TeamGracefulShutdownEvent
- **WHEN** a team undergoes graceful shutdown
- **THEN** a `TeamGracefulShutdownEvent` SHALL be published with TeamID, Reason, BundlesCreated, and MembersSettled

#### Scenario: TeamHealthCheckEvent
- **WHEN** a team-level health sweep completes
- **THEN** a `TeamHealthCheckEvent` SHALL be published with TeamID, Healthy count, and Total count

#### Scenario: TeamMemberUnhealthyEvent
- **WHEN** a member exceeds the missed ping threshold
- **THEN** a `TeamMemberUnhealthyEvent` SHALL be published with TeamID, MemberDID, MemberName, MissedPings, and LastSeenAt

#### Scenario: TeamBudgetWarningEvent
- **WHEN** a team's spend crosses a warning threshold
- **THEN** a `TeamBudgetWarningEvent` SHALL be published with TeamID, Threshold, Spent, and Budget

#### Scenario: TeamLeaderChangedEvent
- **WHEN** a team's leader is replaced
- **THEN** a `TeamLeaderChangedEvent` SHALL be published with TeamID, OldLeaderDID, and NewLeaderDID

## MODIFIED Requirements

### Requirement: Team and Member types
The `p2p/team` package SHALL define `Team`, `Member`, `TeamState`, `MemberRole`, and `MemberStatus` types for representing distributed agent teams.

#### Scenario: Team lifecycle states
- **WHEN** a Team is created
- **THEN** it SHALL progress through states: Forming -> Active -> ShuttingDown -> Disbanded, or Forming -> Active -> Completed

#### Scenario: Member roles
- **WHEN** members join a team
- **THEN** each SHALL have a role: Leader, Worker, Reviewer, or Observer

#### Scenario: Budget fields
- **WHEN** a Team is created
- **THEN** it SHALL have Budget and Spent fields for tracking team expenditure

#### Scenario: MaxMembers enforcement
- **WHEN** a member is added to a team at maximum capacity
- **THEN** AddMember SHALL return ErrTeamFull

### Requirement: TeamCoordinator
The `Coordinator` SHALL provide methods: FormTeam, GetTeam, DelegateTask, CollectResults, DisbandTeam, ListTeams, KickMember, TeamsForMember, GracefulShutdown, LoadPersistedTeams. It SHALL manage the full team lifecycle with optional persistent storage.

#### Scenario: Form team
- **WHEN** FormTeam is called with a list of member DIDs
- **THEN** a new Team SHALL be created, members SHALL be assigned roles, the team SHALL be persisted if a store is configured, and TeamFormedEvent SHALL be published

#### Scenario: Get team
- **WHEN** GetTeam is called with a valid teamID
- **THEN** it SHALL return the team, or ErrTeamNotFound if not found

#### Scenario: Delegate task
- **WHEN** DelegateTask is called with a task description and team ID
- **THEN** the task SHALL be assigned to all workers, TeamTaskDelegatedEvent and TeamTaskCompletedEvent SHALL be published, and the team state SHALL be persisted

#### Scenario: Collect results
- **WHEN** CollectResults is called after task delegation
- **THEN** it SHALL return results from all members that completed their tasks

#### Scenario: Disband team
- **WHEN** DisbandTeam is called
- **THEN** the team state SHALL transition to Disbanded, all members SHALL be released, the team SHALL be deleted from the store, and TeamDisbandedEvent SHALL be published

#### Scenario: List teams
- **WHEN** ListTeams is called
- **THEN** it SHALL return all active teams currently managed by the coordinator

### Requirement: Team configuration
The `P2PConfig` SHALL include a `TeamConfig` struct with health monitoring and membership policy settings.

#### Scenario: Team config fields
- **WHEN** P2PConfig is loaded
- **THEN** it SHALL contain Team.HealthCheckInterval, Team.MaxMissedHeartbeats, and Team.MinReputationScore
