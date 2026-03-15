## Context

The multi-agent orchestration system routes tools to sub-agents using prefix-based matching. Two sources of truth define agent specs: the builtin `agentSpecs` slice in `internal/orchestration/tools.go` and the embedded `AGENT.md` files in `internal/agentregistry/defaults/`. When dynamic specs are loaded (via `agentregistry`), they replace the builtin specs entirely. The Vault AGENT.md was missing prefixes for tool families added after the initial AGENT.md was written (smartaccount, economy, escrow, sentinel, contract), causing those tools to fall through to the "unmatched" bucket.

## Goals / Non-Goals

**Goals:**
- Vault AGENT.md prefixes and keywords match the builtin vault spec exactly
- Builtin vault spec includes all tool families that exist in the codebase
- capabilityMap covers all vault prefixes for diagnostic output

**Non-Goals:**
- Redesigning the dual-source spec system (builtin vs dynamic)
- Adding automated sync validation between AGENT.md and agentSpecs
- Changing the prefix-based routing algorithm

## Decisions

1. **Sync AGENT.md to match builtin spec** -- The AGENT.md is the downstream artifact; when new tool families are added to the builtin spec, the AGENT.md must be updated in lockstep. This fix brings them into alignment.

2. **Add capabilityMap entries for new prefixes** -- The `capabilityMap` drives `builtin_health` diagnostic output. Missing entries cause "general actions" fallback labels, which are unhelpful for diagnostics.

## Risks / Trade-offs

- [Risk] Two sources of truth remain unsynchronized by design -- This fix is a point-in-time sync. Future tool families must update both files. An automated sync check is out of scope but would prevent recurrence.
