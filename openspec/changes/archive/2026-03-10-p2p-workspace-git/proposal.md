# Proposal: P2P Workspace & Git Bundle Integration

## Problem

Lango has a powerful P2P infrastructure (libp2p, GossipSub, DID, ZK, team coordination, payment) but lacks **code sharing (Git)** and **project-level collaboration spaces (Workspace)** — the two features required for agents to co-develop software without a central server.

## Solution

Implement the **Sovereign Swarm** model: decentralized agent workspaces with git-based code sharing, inspired by agenthub's concepts but built on Lango's P2P stack instead of a centralized server.

### Key Components

1. **Workspace Manager** — BoltDB-persisted CRUD for collaborative workspaces with lifecycle (Forming → Active → Archived)
2. **Workspace GossipSub** — Per-workspace messaging via shared GossipSub topics (`/lango/workspace/{id}`)
3. **BareRepoStore** — go-git backed bare repository management per workspace
4. **GitBundleService** — Git bundle create/apply/log/diff/leaves operations
5. **Git Protocol Handler** — libp2p stream handler (`/lango/p2p-git/1.0.0`) for bundle exchange
6. **Chronicler** — Message persistence as graph triples
7. **Contribution Tracker** — Per-member commit/code/message tracking

### Key Design Decisions

- **Git Bundle approach** (like agenthub): Simple bundle exchange instead of full smart protocol
- **Sprawling DAG**: No branches — agents navigate by commit hash
- **Shared PubSub**: Node-level GossipSub instance shared between discovery and workspace
- **Base64 JSON transport**: Consistent with existing A2A protocol
- **Workspace as P2P sub-feature**: Lives under `p2p.workspace.*` config

## Non-Goals

- Full git smart protocol (push/pull/fetch with refs negotiation)
- Merge conflict resolution (sprawling DAG model)
- Central workspace registry/discovery (peer-driven)
- On-chain workspace membership (pure P2P)

## Impact

- 15 new files, 6 modified files
- 2 new packages: `internal/p2p/workspace/`, `internal/p2p/gitbundle/`
- 12 new agent tools, 10 new CLI commands
- 6 new event types
- Zero existing test regressions
