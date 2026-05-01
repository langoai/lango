# Design

## Contract Inventory

- `multi-agent-orchestration`: built-in production path still references legacy transfer compatibility.
- `agent-control-plane-tools`: built-in teammate runtime already exists but is not yet the only production path.
- `agent-routing`: embedded prompt files still require `transfer_to_agent("lango-orchestrator")`.
- `agent-registry`: embedded `AGENT.md` defaults remain part of the production prompt contract.
- `adk-architecture`: `failed to find agent` retry still assumes a useful sub-agent list exists.
- `tool-capability-layer`: grant/recheck semantics must be aligned with the hard cut.
- `run-ledger`: durable visibility expectations must be explicit for built-in teammate runs.
