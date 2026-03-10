# P2P Workspace

## Overview

Collaborative workspaces where P2P agents share code and messages without a central server. Part of the Sovereign Swarm model.

## Package

`internal/p2p/workspace/`

## Types

- **Workspace**: ID, Name, Goal, Status (Forming/Active/Archived), Members, Metadata
- **Member**: DID, Name, Role (creator/member), JoinedAt
- **Message**: ID, Type, WorkspaceID, SenderDID, Content, Metadata, ParentID, Timestamp
- **MessageType**: TASK_PROPOSAL, LOG_STREAM, COMMIT_SIGNAL, KNOWLEDGE_SHARE, MEMBER_JOINED, MEMBER_LEFT

## Components

### Manager
BoltDB-backed workspace lifecycle manager with in-memory cache.
- CRUD: Create, Join, Leave, List, Get, Activate, Archive
- Messaging: Post, Read (with filtering)
- Config: `p2p.workspace.*` (enabled, dataDir, maxWorkspaces, maxBundleSizeBytes, chroniclerEnabled, autoSandbox, contributionTracking)

### WorkspaceGossip
Per-workspace GossipSub topic management using shared PubSub instance from Node.
- Topics: `/lango/workspace/{id}`
- Subscribe/Unsubscribe/Publish/Stop

### Chronicler
Persists workspace messages as graph triples via TripleAdder callback.
- Avoids import cycle with graph store
- Records type, workspace, sender, content, timestamp, replyTo, metadata

### ContributionTracker
In-memory per-member contribution tracking per workspace.
- Tracks: commits, codeBytes, messages, lastActive

## Events

- WorkspaceCreatedEvent, WorkspaceMemberJoinedEvent, WorkspaceMemberLeftEvent
- WorkspaceCommitReceivedEvent, WorkspaceMessagePostedEvent, WorkspaceArchivedEvent

## Agent Tools

- p2p_workspace_create, join, leave, list, status, post, read

## CLI Commands

- `lango p2p workspace create/list/status/join/leave`
