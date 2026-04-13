## Purpose

Capability spec for git-state-tracking. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Git state collection during health pings
The health monitor SHALL collect HEAD commit hashes from team members during successful health pings when a GitStateProvider is configured.

#### Scenario: Collect git state on successful ping
- **WHEN** a health ping succeeds and GitStateProvider and WorkspaceIDsFn are configured
- **THEN** the system queries the member's HEAD hash for each workspace and stores it

#### Scenario: Skip git state when provider not configured
- **WHEN** a health ping succeeds but GitStateProvider is nil
- **THEN** the system does not attempt to collect git state

### Requirement: Git divergence detection
The health monitor SHALL detect when team members have different HEAD commits for the same workspace by comparing against the majority HEAD.

#### Scenario: Detect divergent member
- **WHEN** DetectDivergence is called and one member has a different HEAD than the majority
- **THEN** the system returns a GitDivergence entry with the member's DID, their HEAD, and the majority HEAD

#### Scenario: All members synchronized
- **WHEN** DetectDivergence is called and all members have the same HEAD
- **THEN** the system returns an empty slice

#### Scenario: No members tracked
- **WHEN** DetectDivergence is called for a workspace with no tracked members
- **THEN** the system returns nil

### Requirement: Empty hash ignored
The system SHALL ignore empty HEAD hashes when updating git state.

#### Scenario: Empty hash update is no-op
- **WHEN** updateGitState is called with an empty headHash
- **THEN** the system does not store the entry
