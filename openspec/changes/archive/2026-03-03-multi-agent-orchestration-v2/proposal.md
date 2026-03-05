## Why

The current multi-agent system uses hardcoded AgentSpec definitions with static prefix-based tool routing. To support dynamic agent extension, semantic routing, context isolation, tool lifecycle hooks, and P2P distributed agent teams, the orchestration layer needs a declarative registry, hook infrastructure, sub-session isolation, and agent pool coordination.

## What Changes

- **Agent Registry**: Declarative AGENT.md files (YAML frontmatter + markdown body) replace hardcoded `agentSpecs`. Embedded defaults via `embed.FS`, user-defined agents via `~/.lango/agents/`, with override semantics (User > Embedded > Builtin).
- **Dynamic Tool Partitioning**: `DynamicToolSet` and `PartitionToolsDynamic()` use registry specs instead of hardcoded prefix maps. Existing `PartitionTools()` preserved as wrapper for backward compatibility.
- **Description-Based Routing**: Routing table entries include `Capabilities` field for semantic matching alongside existing keyword routing.
- **Tool Execution Hooks**: `PreToolHook`/`PostToolHook` interfaces with priority-based `HookRegistry`. Built-in hooks: SecurityFilter, AgentAccessControl, EventBus, KnowledgeSave. `WithHooks()` middleware bridges into existing toolchain.
- **Agent Name Context Propagation**: `WithAgentName(ctx)`/`AgentNameFromContext(ctx)` for hook and middleware agent identification.
- **Sub-Session & Context Isolation**: `ChildSession` with "read parent, write child" semantics. `StructuredSummarizer` (zero-cost default) and `LLMSummarizer` (opt-in).
- **Agent Memory**: In-memory agent-scoped persistent memory store with save/recall/forget tools. Scope resolution: instance > type > global.
- **P2P Agent Pool**: Dynamic remote agent management with weighted scoring (trust 0.35, capability 0.25, performance 0.20, price 0.15, availability 0.05).
- **P2P Team Coordination**: `TeamCoordinator` for forming teams, delegating tasks, collecting results, disbanding. Conflict resolution strategies: TrustWeighted, MajorityVote, LeaderDecides, FailOnConflict.
- **P2P Team Payment**: Trust-based payment negotiation integrated with existing PayGate/Settlement services.
- **P2P Dynamic Agent Provider**: `DynamicAgentProvider` interface wired into orchestrator routing table for P2P agent discovery.
- **CLI Updates**: `lango agent list` shows dynamic registry (builtin/embedded/user/remote). `lango agent status` includes registry counts, P2P, and hooks status.

## Capabilities

### New Capabilities
- `agent-registry`: Declarative AGENT.md-based agent definition, parsing, registry with override semantics, embedded defaults, and file store
- `tool-execution-hooks`: PreToolUse/PostToolUse hook system with priority-based registry and middleware bridge
- `sub-session-isolation`: Child session forking, merge, discard with structured summarization for sub-agent context isolation
- `agent-memory`: Agent-scoped persistent memory with save/recall/forget tools and scope resolution
- `p2p-agent-pool`: Dynamic remote agent pool management with weighted scoring and health checking
- `p2p-team-coordination`: Distributed agent team formation, task delegation, result collection, and conflict resolution
- `p2p-team-payment`: Trust-based payment negotiation for P2P team task delegation
- `agent-context-propagation`: Agent name injection into Go context for hook and middleware identification

### Modified Capabilities
- `multi-agent-orchestration`: Dynamic specs support via `Config.Specs`, `DynamicAgents` provider, capability-based routing enhancement
- `cli-agent-inspection`: Registry-aware agent list/status with builtin/embedded/user/remote source display

## Impact

- **Core packages**: New packages `agentregistry`, `agentmemory`, `ctxkeys`, `toolchain` (hooks), `session` (child), `p2p/agentpool`, `p2p/team`
- **Modified packages**: `orchestration` (dynamic specs, P2P provider), `app` (wiring), `cli/agent` (registry-aware), `config` (new config types), `eventbus` (team events), `p2p/protocol` (team messages), `adk` (child session, context)
- **Config additions**: `agent.agentsDir`, `hooks.enabled`, `p2p.team.*`
- **No breaking changes**: All existing APIs preserved; new functionality is additive
