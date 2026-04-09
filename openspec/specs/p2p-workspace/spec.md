## Purpose

Capability spec for p2p-workspace. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: p2p-workspace capability documented
The p2p-workspace capability SHALL be documented through the sections in this spec. This requirement is a structural placeholder that satisfies the canonical openspec format; detailed behavior contracts live in the descriptive sections of this file.

#### Scenario: Spec file is readable
- **WHEN** the p2p-workspace spec.md file is read
- **THEN** it SHALL describe the capability's behavior in sections below

# P2P Workspace

## Overview

Collaborative workspaces where P2P agents share code and messages without a central server. Part of the Sovereign Swarm model.

## Package

`internal/p2p/workspace/`

## Types

- **Workspace**: ID, Name, Goal, Status (Forming/Active/Archived), Members, Metadata
- **Member**: DID, Name, Role (`Role` type: RoleCreator/RoleMember), JoinedAt
- **Message**: ID, Type, WorkspaceID, SenderDID, Content, Metadata, ParentID, Timestamp
- **MessageType**: TASK_PROPOSAL, LOG_STREAM, COMMIT_SIGNAL, KNOWLEDGE_SHARE, MEMBER_JOINED, MEMBER_LEFT, CONFLICT_REPORT, BRANCH_CREATED, BRANCH_MERGED, SYNC_REQUEST

## Components

### Manager
BoltDB-backed workspace lifecycle manager with in-memory cache.
- CRUD: Create, Join, Leave, List, Get, Activate, Archive
- Messaging: Post, Read (with filtering)
- Config: `p2p.workspace.*` (enabled, dataDir, maxWorkspaces, maxBundleSizeBytes, chroniclerEnabled, autoSandbox, contributionTracking)

### WorkspaceGossip
Per-workspace GossipSub topic management using shared PubSub instance from Node (`sync.Once`-protected).
- Topics: `/lango/workspace/{id}`
- Subscribe/Unsubscribe/Publish/Stop
- Constructed once with message handler pre-configured

### Chronicler
Persists workspace messages as graph triples via TripleAdder callback.
- Avoids import cycle with graph store
- Records type, workspace, sender, content, timestamp, replyTo, metadata

### ContributionTracker
In-memory per-member contribution tracking per workspace.
- Tracks: commits, codeBytes, messages, lastActive
- Remove: cleanup data for a workspace

## Events

- WorkspaceCreatedEvent, WorkspaceMemberJoinedEvent, WorkspaceMemberLeftEvent
- WorkspaceCommitReceivedEvent, WorkspaceMessagePostedEvent, WorkspaceArchivedEvent
- WorkspaceGitDivergenceEvent

## Branch Collaboration Messages

### Requirement: Branch collaboration message types
The workspace message system SHALL support four new message types for branch-based collaboration signaling: CONFLICT_REPORT, BRANCH_CREATED, BRANCH_MERGED, SYNC_REQUEST.

#### Scenario: Conflict report message
- **WHEN** a merge conflict occurs between branches
- **THEN** a CONFLICT_REPORT message is posted with metadata containing conflictFiles, sourceBranch, targetBranch, sourceAgent, taskID, and resolution fields

#### Scenario: Branch created message
- **WHEN** a task branch is created in a workspace
- **THEN** a BRANCH_CREATED message is posted to notify other workspace members

#### Scenario: Branch merged message
- **WHEN** a task branch is successfully merged
- **THEN** a BRANCH_MERGED message is posted to notify other workspace members

#### Scenario: Sync request message
- **WHEN** git state divergence is detected or a member requests synchronization
- **THEN** a SYNC_REQUEST message is posted to coordinate re-sync

## Agent Tools

- p2p_workspace_create, join, leave, list, status, post, read

## CLI Commands

- `lango p2p workspace create/list/status/join/leave`
