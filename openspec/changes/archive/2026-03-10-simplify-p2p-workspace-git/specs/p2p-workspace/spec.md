## MODIFIED Requirements

### Types
- **Member**: DID, Name, Role (`Role` type: RoleCreator/RoleMember), JoinedAt

### ContributionTracker
In-memory per-member contribution tracking per workspace.
- Tracks: commits, codeBytes, messages, lastActive
- Remove: cleanup data for a workspace

## ADDED Requirements

### Requirement: Typed Member Role
Member.Role SHALL use the `Role` string type with constants `RoleCreator` and `RoleMember` instead of raw strings.

#### Scenario: Creator role assignment
- **WHEN** a workspace is created
- **THEN** the creator member SHALL have Role set to `RoleCreator`

#### Scenario: Member role assignment
- **WHEN** an agent joins an existing workspace
- **THEN** the new member SHALL have Role set to `RoleMember`

### Requirement: ContributionTracker workspace cleanup
ContributionTracker SHALL provide a `Remove(workspaceID)` method that deletes all contribution data for a workspace.

#### Scenario: Remove workspace contributions
- **WHEN** `Remove(workspaceID)` is called
- **THEN** all contribution data for that workspace SHALL be deleted from the tracker

### Requirement: Thread-safe PubSub initialization
`Node.PubSub()` SHALL use `sync.Once` to guarantee exactly one GossipSub instance is created per node, even under concurrent access.

#### Scenario: Concurrent PubSub access
- **WHEN** multiple goroutines call `Node.PubSub()` concurrently
- **THEN** exactly one GossipSub instance SHALL be created and shared

### Requirement: Single WorkspaceGossip construction
WorkspaceGossip SHALL be constructed exactly once with the message handler already configured, not constructed and then replaced.

#### Scenario: Gossip initialization
- **WHEN** workspace is initialized with chronicler and tracker enabled
- **THEN** WorkspaceGossip SHALL be created once with the handler that dispatches to both chronicler and tracker
