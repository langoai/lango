## Why

The lango project has undergone major internal feature development (multi-agent orchestration, P2P teams, ZKP, learning system, event bus, agent registry, tool hooks, approval system, etc.) but many of these features lack CLI/TUI exposure and documentation. Users cannot inspect, configure, or manage these features without direct code access. Closing these gaps is necessary to make the platform usable and discoverable.

## What Changes

- Add P2P Teams CLI (`p2p team list/status/disband`) for managing peer-to-peer team lifecycle
- Add ZKP CLI (`p2p zkp status/circuits`) for inspecting zero-knowledge proof configuration
- Add Agent CLI enhancements (`agent tools`, `agent hooks`) for tool catalog and hook inspection
- Add A2A CLI (`a2a card`, `a2a check <url>`) for agent-to-agent protocol management
- Add Learning CLI (`learning status`, `learning history`) for learning system inspection
- Add Librarian CLI (`librarian status`, `librarian inquiries`) for librarian monitoring
- Add Approval CLI (`approval status`) for approval system dashboard
- Add Graph CLI extensions (`graph add`, `graph export`, `graph import`) for knowledge graph manipulation
- Add Memory CLI (`memory agents`, `memory agent <name>`) for agent memory inspection
- Add Payment CLI (`payment x402`) for X402 protocol configuration display
- Add Workflow CLI (`workflow validate <file>`) for YAML workflow validation
- Add TUI Settings forms for hooks configuration and agent memory
- Enhance TUI Settings with additional multi-agent fields (maxDelegationRounds, maxTurns, errorCorrectionEnabled, agentsDir)
- Enhance TUI Settings with additional librarian/knowledge fields
- Add 4 new doctor health checks (tool hooks, agent registry, librarian, approval)
- Add onboard hints for advanced features (agent memory, hooks, librarian)
- Add comprehensive feature documentation (agent-format, learning, ZKP, approval)
- Update CLI reference docs, README, and index

## Capabilities

### New Capabilities
- `cli-p2p-teams`: CLI commands for P2P team lifecycle management (list, status, disband)
- `cli-zkp-inspection`: CLI commands for ZKP system inspection (status, circuits)
- `cli-agent-tools-hooks`: CLI commands for tool catalog listing and hook configuration display
- `cli-a2a-management`: CLI commands for A2A agent card display and remote card fetching
- `cli-learning-inspection`: CLI commands for learning system status and history
- `cli-librarian-monitoring`: CLI commands for librarian status and inquiry listing
- `cli-approval-dashboard`: CLI command for approval system status display
- `cli-graph-extended`: CLI commands for graph triple add, export (JSON/CSV), and import
- `cli-agent-memory`: CLI commands for agent memory inspection
- `cli-x402-config`: CLI command for X402 protocol configuration display
- `cli-workflow-validate`: CLI command for YAML workflow validation without execution
- `tui-hooks-settings`: TUI settings form for hooks configuration
- `tui-agent-memory-settings`: TUI settings form for agent memory configuration

### Modified Capabilities
- `cli-graph-management`: Adding add/export/import subcommands + AllTriples interface method
- `cli-memory-management`: Adding agent memory subcommands + ListAgentNames/ListAll interface methods
- `cli-payment-management`: Adding x402 subcommand
- `cli-workflow-management`: Adding validate subcommand
- `cli-p2p-management`: Adding team and zkp subcommand groups
- `cli-doctor`: Adding 4 new health checks (tool hooks, agent registry, librarian, approval)
- `cli-health-check`: Adding advanced feature hints to onboard flow

## Impact

- **Code**: 32 new files, 28 modified files across `internal/cli/`, `internal/graph/`, `internal/agentmemory/`, `cmd/lango/`
- **Interfaces**: `graph.Store` gains `AllTriples()`, `agentmemory.Store` gains `ListAgentNames()`/`ListAll()`
- **CLI**: 18 new subcommands registered across 11 command groups
- **TUI**: 2 new settings forms (hooks, agent memory), 15+ new form fields
- **Docs**: 6 new doc files, 12 modified doc files, README updated
- **Dependencies**: No new external dependencies; all features built on existing internal packages
