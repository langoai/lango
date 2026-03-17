## Context

The core EntryPoint v0.7 migration and CirclePermitProvider were implemented but downstream artifacts (CLI commands, TUI settings forms, documentation) were not updated in the same pass. This change is a pure sync pass — no new architecture or logic.

## Goals / Non-Goals

**Goals:**
- Replace all v0.6 EntryPoint address references with v0.7
- Expose `mode` field in CLI status output, TUI forms, and documentation
- Add permit mode support to CLI's `initPaymasterProvider()`

**Non-Goals:**
- No new features or behavior changes
- No changes to core/internal packages (already done)

## Decisions

### 1. Global replace for EntryPoint address

**Decision**: `replace_all` across all files referencing the v0.6 address.

**Rationale**: The address is a constant — no conditional logic needed. All references should point to v0.7.

### 2. CLI deps mirrors app wiring pattern

**Decision**: CLI `initPaymasterProvider()` updated with the same permit mode branching as `app/wiring_smartaccount.go`.

**Rationale**: CLI commands need to create the same provider types as the full app, just without recovery wrapping.

## Risks / Trade-offs

- **[None]** — This is a mechanical sync pass with no behavioral changes.
