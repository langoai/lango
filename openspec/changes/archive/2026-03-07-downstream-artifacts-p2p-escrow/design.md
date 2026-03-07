## Context

The `feature/p2p-escrow` branch shipped major code changes (on-chain escrow Hub/Vault, Security Sentinel, P2P Settlement, Team Coordination enhancements) without updating downstream artifacts. Users and agents cannot discover or configure these features without accurate docs, prompts, and TUI surfaces. This change synchronizes all user-facing documentation and configuration UI with the implemented code.

## Goals / Non-Goals

**Goals:**
- Update all downstream artifacts to accurately reflect the implemented p2p-escrow features
- Add TUI form for on-chain escrow configuration (the only code change)
- Ensure consistency between code, docs, CLI docs, config docs, system prompts, and README

**Non-Goals:**
- No new feature implementation — all features are already coded
- No changes to core business logic, APIs, or data models
- No migration or deployment changes required

## Decisions

1. **TUI Form Pattern**: Follow existing `NewEconomyEscrowForm()` pattern in `forms_economy.go` for the new `NewEconomyEscrowOnChainForm()`. Rationale: consistency with established codebase patterns, reuses `tuicore.FormModel` and `tuicore.Field` types.

2. **Documentation Structure**: Update existing doc files rather than creating new ones. On-chain escrow docs go into `economy.md` (subsection), sentinel into same file, contracts into `contracts.md`. Rationale: keeps related features co-located, avoids doc fragmentation.

3. **Tool Name Convention**: Use the actual registered tool names (`escrow_*`, `sentinel_*`) rather than the old `economy_escrow_*` prefix in all documentation. Rationale: matches `internal/app/tools_escrow.go` and `tools_sentinel.go` registrations.

4. **Parallel Documentation Work**: Split 9 work units across 3 parallel agents + lead for maximum throughput. Rationale: documentation units are independent, no cross-dependencies between WUs.

## Risks / Trade-offs

- [Risk] Documentation drift if code changes after docs are written → Mitigation: docs reference source-of-truth files; OpenSpec archive captures the snapshot
- [Risk] TUI form fields may not cover all config options → Mitigation: form mirrors `types_economy.go` `EscrowOnChainConfig` + `EscrowSettlementConfig` structs exactly (10 fields)
