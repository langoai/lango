# Design: P2P Workspace & Git Bundle Integration

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  App Layer (internal/app/)                                   │
│  ┌──────────────────┐  ┌──────────────────┐                 │
│  │ tools_workspace  │  │ wiring_workspace │                 │
│  │  (12 agent tools)│  │  (init + wire)   │                 │
│  └───────┬──────────┘  └───────┬──────────┘                 │
│          │                     │                             │
├──────────┼─────────────────────┼─────────────────────────────┤
│  P2P Layer                     │                             │
│  ┌─────────────────────────────┼───────────────────────┐     │
│  │  workspace/                 │                       │     │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐            │     │
│  │  │ Manager  │ │ Gossip   │ │Chronicler│            │     │
│  │  │ (BoltDB) │ │(PubSub)  │ │(Triples) │            │     │
│  │  └──────────┘ └──────────┘ └──────────┘            │     │
│  │  ┌──────────────────────────────────┐              │     │
│  │  │ ContributionTracker (in-memory)  │              │     │
│  │  └──────────────────────────────────┘              │     │
│  └────────────────────────────────────────────────────┘     │
│  ┌────────────────────────────────────────────────────┐     │
│  │  gitbundle/                                         │     │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐            │     │
│  │  │BareRepo  │ │ Service  │ │ Handler  │            │     │
│  │  │Store     │ │(bundle)  │ │(protocol)│            │     │
│  │  └──────────┘ └──────────┘ └──────────┘            │     │
│  └────────────────────────────────────────────────────┘     │
│                                                              │
│  Shared: Node.PubSub(), SessionStore, Firewall, EventBus    │
└──────────────────────────────────────────────────────────────┘
```

## Key Design Decisions

### 1. Shared GossipSub Instance
Node.PubSub() creates a single GossipSub instance lazily, shared between:
- Discovery service (agent card propagation)
- Workspace gossip (per-workspace messaging)

This avoids libp2p's limitation of one PubSub per host.

### 2. Callback Pattern for Chronicler
Chronicler uses `TripleAdder func(ctx, []Triple) error` callback instead of importing graph.Store directly. This avoids import cycles between workspace and graph packages.

### 3. Git Bundle over Smart Protocol
Uses `git bundle create/unbundle` CLI commands for simplicity. go-git is used for programmatic access (log, refs, commit traversal) but bundles require the git CLI.

### 4. Session Validation for Git Protocol
Git protocol handler receives a SessionValidator callback that iterates over active sessions to find matching tokens, avoiding a direct dependency on handshake.SessionStore.

### 5. Workspace as P2P Sub-Feature
Workspace initializes inside initP2P flow, sharing the P2P node, sessions, and firewall. Config lives under `p2p.workspace.*`.

## Package Dependencies

```
app/ → workspace/, gitbundle/ (tools + wiring)
workspace/ → bbolt, pubsub, zap (no app imports)
gitbundle/ → go-git, libp2p/network, zap (no app imports)
eventbus/ → time (standalone events)
cli/p2p/ → bootstrap, config (lazy loading)
```

## Data Storage

- **Workspace metadata**: BoltDB (`~/.lango/workspaces/workspaces.db`)
- **Workspace messages**: Same BoltDB, separate bucket
- **Git repos**: Bare repos at `~/.lango/workspaces/{id}/repo.git`
- **Contributions**: In-memory only (reset on restart)
