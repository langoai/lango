# Dynamic Multi-Agent Hard Cut

## Problem

Built-in multi-agent execution still mixes two incompatible models:

1. the control-plane teammate runtime based on `agent_spawn`
2. legacy ADK static delegation based on `transfer_to_agent`

This ambiguity leaks into prompts, embedded `AGENT.md` files, skill execution guidance, hallucinated-agent recovery, and operator surfaces. It also leaves capability-runtime and RunLedger observability gaps unresolved.

## Proposed Change

Make built-in teammate execution spawn-only in production. Remove built-in reliance on `transfer_to_agent`, keep remote A2A separate, narrow ADK recovery to the remaining remote/legacy transfer surface, tighten capability-runtime blocked-call behavior, and update downstream docs and CLI surfaces to match the new contract.

## User-Facing Impact

Built-in teammate work routes through `agent_spawn`-backed execution only. Remote A2A remains available as a separate compatibility path. Copied custom `AGENT.md` files that still encode built-in `transfer_to_agent("lango-orchestrator")` behavior require upgrade guidance.
