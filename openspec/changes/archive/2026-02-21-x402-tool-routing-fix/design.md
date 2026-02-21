## Context

`PartitionTools()` in `internal/orchestration/tools.go` routes tools to sub-agents by matching tool name prefixes. The vault agent owns prefixes `crypto_`, `secrets_`, and `payment_`. The `x402_fetch` tool was named without the `payment_` prefix, so it falls into the Unmatched bucket and is never assigned to any sub-agent. This is the only tool (out of 71) with this routing bug.

## Goals / Non-Goals

**Goals:**
- Fix the tool routing bug so `x402_fetch` is correctly assigned to the vault agent
- Maintain consistency with the `payment_` prefix naming convention used by all other payment tools

**Non-Goals:**
- Changing the `PartitionTools` routing logic itself
- Adding fallback/fuzzy matching for unmatched tools
- Modifying archive records of past changes

## Decisions

**Decision: Rename tool to `payment_x402_fetch` (not add special-case routing)**

Rationale: All other payment tools (`payment_send`, `payment_balance`, `payment_history`, `payment_limits`, `payment_wallet_info`, `payment_create_wallet`) follow the `payment_` prefix convention. Renaming the tool to match the convention is the simplest, most consistent fix. Adding a special case in the routing logic would add complexity and deviate from the established pattern.

Alternatives considered:
- Add `x402_` as a vault prefix — introduces a one-off prefix for a single tool
- Add special-case routing in `PartitionTools` — adds complexity, breaks the clean prefix model

## Risks / Trade-offs

- [Risk] External references to `x402_fetch` tool name → No external consumers exist; tool is internal to the agent system, so no migration needed.
- [Risk] Archive docs reference old name → Accepted; archive records are historical and MUST NOT be modified.
