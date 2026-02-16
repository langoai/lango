## Why

Graph Store, Multi-Agent Orchestration, A2A Protocol, Graph RAG, and Self-Learning Graph features are implemented internally but not exposed to users. There are no CLI commands, no doctor checks, no onboard wizard screens, and no README documentation for these features. Users cannot configure, monitor, or understand these capabilities.

## What Changes

- Add `Count`, `PredicateStats`, `ClearAll` methods to the graph `Store` interface and `BoltStore` implementation
- Add `lango graph` CLI commands: `status`, `query`, `stats`, `clear`
- Add `lango agent` CLI commands: `status`, `list`
- Add doctor checks for Graph Store, Multi-Agent, and A2A configurations
- Add onboard wizard screens for Graph Store, Multi-Agent, and A2A Protocol
- Add config defaults and validation for Graph and A2A settings
- Update README with documentation for all new features and CLI commands

## Capabilities

### New Capabilities
- `cli-graph-management`: CLI commands to inspect, query, and manage the knowledge graph store
- `cli-agent-inspection`: CLI commands to inspect agent mode, sub-agents, and remote A2A agents

### Modified Capabilities
- `graph-store`: Add Count, PredicateStats, ClearAll methods to Store interface
- `cli-doctor`: Add GraphStoreCheck, MultiAgentCheck, A2ACheck diagnostic checks
- `cli-onboard`: Add Graph Store, Multi-Agent, and A2A Protocol wizard screens and forms
- `config-system`: Add Graph and A2A defaults to DefaultConfig and validation to Validate

## Impact

- `internal/graph/store.go` — Store interface extended with 3 new methods
- `internal/graph/bolt_store.go` — BoltStore implementation of new methods
- `internal/cli/graph/` — New package (5 files)
- `internal/cli/agent/` — New package (3 files)
- `internal/cli/doctor/checks/` — 3 new check files + AllChecks() registration
- `internal/cli/onboard/` — menu.go, forms_impl.go, wizard.go, state_update.go modified
- `internal/config/loader.go` — DefaultConfig and Validate updated
- `cmd/lango/main.go` — Wire graph and agent commands
- `internal/learning/graph_engine_test.go` — fakeGraphStore updated for new interface
- `README.md` — New sections, CLI commands, config reference, architecture tree
