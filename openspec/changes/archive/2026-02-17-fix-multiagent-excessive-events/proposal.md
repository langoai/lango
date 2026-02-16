## Why

Multi-agent orchestration (`agent.multiAgent: true`) generates 18+ agent events for a single simple message like "hello!", with repeated "Event from an unknown agent: lango-agent" warnings and severely degraded response time. Two bugs (author mismatch, A2A wiring order) and three optimization gaps (no short-circuit for simple queries, unconditional sub-agent creation, no delegation round limit) are the root causes.

## What Changes

- Fix EventsAdapter author mapping to use the actual root agent name (`lango-orchestrator`) instead of hardcoded `lango-agent`, eliminating "unknown agent" warnings
- Add `Author` field to `session.Message` and ent schema for persisting agent identity across sessions
- Fix A2A remote agent loading order so remote agents are included in the agent tree
- Add orchestrator short-circuit instruction allowing direct responses to simple conversational queries
- Make sub-agent creation conditional based on whether tools are actually assigned to each role
- Add `MaxDelegationRounds` config to limit orchestrator delegation depth per user turn

## Capabilities

### New Capabilities

(none)

### Modified Capabilities
- `multi-agent-orchestration`: Fix author mismatch bug, add conditional sub-agent creation, add short-circuit for simple queries, add max delegation rounds
- `a2a-protocol`: Fix remote agent loading order so agents are included before tree construction
- `ent-session-store`: Add `author` field to message schema for multi-agent identity persistence
- `session-store`: Add `Author` field to `Message` struct

## Impact

- `internal/session/store.go` - Message struct gains Author field
- `internal/ent/schema/message.go` - New author column (requires `go generate`)
- `internal/adk/state.go` - EventsAdapter and SessionAdapter gain rootAgentName
- `internal/adk/session_service.go` - SessionServiceAdapter gains rootAgentName, AppendEvent saves author
- `internal/adk/agent.go` - Agent name passed to session service
- `internal/app/wiring.go` - A2A loading moved before BuildAgentTree, MaxDelegationRounds set
- `internal/orchestration/orchestrator.go` - Conditional sub-agent creation, prompt optimization, max rounds config
