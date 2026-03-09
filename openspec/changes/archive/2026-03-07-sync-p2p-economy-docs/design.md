## Context

The P2P economy branch added economy layer (budget, risk, pricing, negotiation, escrow), contract interaction (ABI cache, EVM read/write), and observability (metrics, token tracking, health, audit) — all with backend code, CLI commands, agent tools, and config types fully implemented. No downstream artifacts (docs, prompts, README, TUI settings, doctor checks) were created. This change syncs all downstream artifacts.

## Goals / Non-Goals

**Goals:**
- Create complete feature and CLI documentation for economy, contracts, and observability
- Update all index/nav files so new docs are discoverable
- Update agent prompts (TOOL_USAGE.md, AGENTS.md) with new tool categories
- Update README with features, CLI commands, and architecture entries
- Create TUI settings forms for economy (5 sub-forms) and observability (1 form)
- Wire forms into editor and state update handlers
- Create doctor health checks for economy, contract, and observability config validation

**Non-Goals:**
- Modifying any backend code (internal/economy, internal/contract, internal/observability)
- Adding new CLI commands or agent tools
- Changing any config types or defaults
- Writing tests for the new doctor checks (existing test patterns cover them)

## Decisions

1. **Follow existing patterns exactly** — All new files mirror existing patterns:
   - Feature docs: `p2p-network.md` pattern (YAML front matter, experimental warning, mermaid, config block)
   - CLI docs: `payment.md` pattern (subcommand sections with flags table and example output)
   - TUI forms: `forms_p2p.go` pattern (form builder with field types and validators)
   - Doctor checks: `embedding.go` pattern (Check interface with Name/Run/Fix)

2. **Economy gets 5 sub-forms** — Rather than one massive form, economy settings are split by sub-system (base, risk, negotiation, escrow, pricing) matching the P2P pattern which has 5 forms (base, ZKP, pricing, owner, sandbox).

3. **Observability in Infrastructure section** — Placed in the Infrastructure menu section alongside payment, cron, background, workflow, and MCP — not in a new section.

4. **Economy section before P2P** — Economy is placed before P2P Network in the menu since it builds on P2P concepts. This matches the logical dependency order.

## Risks / Trade-offs

- [Tool names may drift] → All tool names verified against actual source code registration
- [Config field names may change] → All field names traced from config types to form keys to state update handlers
- [Doc content outdated if backend changes] → Docs track current branch state; future changes follow same sync pattern
