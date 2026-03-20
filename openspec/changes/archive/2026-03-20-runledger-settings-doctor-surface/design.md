## Context

RunLedger is already exposed through runtime config, `lango run ...` CLI commands, doctor-adjacent status output, and documentation. However, `lango settings` still has no RunLedger category, and `lango doctor` has no RunLedger-specific validation. That gap is now visible because RunLedger is no longer a scaffolded experiment: it has enabled flags, rollout controls, timeouts, retention, and workspace isolation settings that operators need to inspect and edit safely.

Relevant code:

- `internal/cli/settings/menu.go` — category registration
- `internal/cli/settings/setup_flow.go` — category → form routing
- `internal/cli/settings/forms_automation.go` — automation-adjacent forms
- `internal/cli/tuicore/state_update.go` — form field → config mapping
- `internal/cli/doctor/checks/checks.go` — diagnostic check registry
- `internal/cli/doctor/doctor.go` — doctor command help text
- `internal/config/types_runledger.go` — RunLedger config surface

## Goals / Non-Goals

**Goals**

- Add a RunLedger category to the settings TUI
- Expose all `RunLedgerConfig` fields through a dedicated form
- Persist RunLedger settings through the existing `ConfigState.UpdateConfigFromForm` path
- Add a doctor check that validates enabled-state invariants and value sanity
- Keep help text and docs aligned with the new settings/doctor surface

**Non-Goals**

- Changing RunLedger runtime behavior
- Adding new RunLedger config fields
- Adding a dedicated TUI status dashboard for RunLedger runs
- Auto-fixing doctor issues beyond guidance-level output

## Decisions

### Decision 1: Place RunLedger under the Automation section

RunLedger sits closest to Cron, Background, and Workflow in operator mental models: it governs durable task execution and rollout controls. Adding it under `Automation` keeps the section coherent and avoids inventing a new singleton section.

### Decision 2: Expose the full existing RunLedgerConfig surface

The form will include:

- `enabled`
- `shadow`
- `writeThrough`
- `authoritativeRead`
- `workspaceIsolation`
- `staleTtl`
- `maxRunHistory`
- `validatorTimeout`
- `plannerMaxRetries`

This matches the config type exactly. No hidden advanced-only sub-surface is added in this change.

### Decision 3: Keep settings form implementation in `forms_automation.go`

RunLedger belongs with Cron/Background/Workflow from a user-facing configuration perspective. Keeping the new form in `forms_automation.go` avoids fragmenting form ownership and follows the existing package structure.

### Decision 4: RunLedger doctor check validates configuration invariants, not environment state

The doctor check will focus on config semantics:

- Skip if RunLedger is disabled
- Fail when `staleTtl <= 0`
- Fail when `validatorTimeout <= 0`
- Fail when `maxRunHistory < 0`
- Fail when `plannerMaxRetries < 0`
- Fail when `authoritativeRead == true` and `writeThrough == false`
- Pass otherwise

This keeps the check deterministic and aligned with what doctor can reliably inspect from config alone.

### Decision 5: Doctor help text should stop pretending counts are fixed

The current doctor long description hardcodes a stale count. This change updates the text to enumerate categories/check families without asserting a fixed total, and explicitly includes RunLedger diagnostics.

## Risks / Trade-offs

- **Risk**: Putting RunLedger in `Automation` may not match every user's intuition.
  - **Mitigation**: It is the least disruptive placement and consistent with existing task-execution categories.
- **Risk**: Doctor may flag `authoritativeRead` without `writeThrough` as too strict.
  - **Mitigation**: This is an actual operator hazard because authoritative reads without ledger-fed write paths can surface incomplete state.
- **Trade-off**: `plannerMaxRetries` is exposed in settings even though current runtime use is limited.
  - **Accepted**: The field already exists in config and is surfaced elsewhere (`lango run status`), so hiding it in settings would be more confusing.
