# P2P Workspace Spec

## Overview

Collaborative workspaces where P2P agents share code and messages without a central server.

## Requirements

### Workspace Lifecycle
- Workspaces have three states: Forming → Active → Archived
- Creator automatically becomes a member with "creator" role
- Max workspaces configurable (default 10)
- BoltDB persistence for workspace metadata and messages

### Workspace Messaging
- Six message types: TASK_PROPOSAL, LOG_STREAM, COMMIT_SIGNAL, KNOWLEDGE_SHARE, MEMBER_JOINED, MEMBER_LEFT
- Messages broadcast via per-workspace GossipSub topics (`/lango/workspace/{id}`)
- Shared GossipSub instance between discovery and workspace (Node.PubSub())
- Message filtering by type, sender, time range, parent ID

### Git Bundle Exchange
- Bare git repositories managed per workspace via go-git
- Git bundles for code sharing (create/apply via CLI, no smart protocol)
- Sprawling DAG model — no branches, navigate by commit hash
- DAG leaf detection for finding latest work
- Bundle size limit configurable (default 50MB)
- Protocol: `/lango/p2p-git/1.0.0` over libp2p streams

### Chronicler
- Persists workspace messages as graph triples (subject-predicate-object)
- Uses callback pattern to avoid import cycles with graph store
- Records message metadata: type, workspace, sender, content, timestamp, replies

### Contribution Tracking
- Tracks per-member: commit count, code bytes, message count, last active time
- In-memory tracking with thread-safe concurrent access
- Per-workspace granularity

### Configuration
- Config path: `p2p.workspace.*`
- Fields: enabled, dataDir, maxWorkspaces, maxBundleSizeBytes, chroniclerEnabled, autoSandbox, contributionTracking

### Agent Tools (12)
- Workspace: create, join, leave, list, status, post, read
- Git: init, push, log, diff, leaves

### CLI Commands (10)
- `lango p2p workspace create/list/status/join/leave`
- `lango p2p git init/log/diff/push/fetch`

### Events (6)
- WorkspaceCreatedEvent, WorkspaceMemberJoinedEvent, WorkspaceMemberLeftEvent
- WorkspaceCommitReceivedEvent, WorkspaceMessagePostedEvent, WorkspaceArchivedEvent
