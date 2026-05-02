# Built-In Hard Cut Enforcement

## Why

The earlier hard-cut work changed prompts and runtime guidance, but it did not complete the structural cut in production code. `BuildAgentTree()` still attaches built-in teammate specs as ADK sub-agents, which keeps `transfer_to_agent` available and leaves hallucinated transfer targets such as `researcher` or `web_search` possible.

## What Changes

This corrective change enforces the hard cut in production:

- keep built-in teammate entries in the routing table
- stop attaching built-in teammate specs as production ADK sub-agents
- keep remote A2A agents attached as sub-agents
- allow built-in hallucinated-target correction even when the orchestrator has zero built-in ADK sub-agents
- add production-wiring regression tests for built-in-only and remote-only tree composition

## User Impact

Built-in teammate work remains routed through `agent_spawn`, but this change makes the production ADK tree match that contract. Remote A2A remains attached through the existing sub-agent compatibility path.
