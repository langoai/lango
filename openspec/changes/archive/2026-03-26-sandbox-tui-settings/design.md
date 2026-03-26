## Context

The sandbox TUI form follows the established pattern from `forms_p2p.go`, `forms_security.go`. The key design decision is namespace separation: P2P sandbox uses `sandbox_*` keys while the new general sandbox uses `os_sandbox_*` to prevent `UpdateConfigFromForm()` from writing to the wrong config path.

## Goals / Non-Goals

**Goals:**
- Make all `cfg.Sandbox.*` fields configurable via TUI settings
- Prevent any cross-contamination with P2P sandbox config

**Non-Goals:**
- Changing the config schema
- Adding dependencies/prerequisites for the sandbox category

## Decisions

### 1. `os_sandbox_*` field key prefix
**Choice**: All field keys prefixed with `os_sandbox_` instead of reusing `sandbox_`.
**Rationale**: P2P sandbox already uses `sandbox_enabled`, `sandbox_timeout` in `state_update.go:517-548`. Same keys would route new form values to `cfg.P2P.ToolIsolation` instead of `cfg.Sandbox`.

### 2. Always-enabled category (no prerequisites)
**Choice**: `categoryIsEnabled()` returns `true` unconditionally for `os_sandbox`.
**Rationale**: OS sandbox is an independent feature — unlike P2P sandbox which requires `cfg.P2P.Enabled`, the general sandbox has no parent toggle.

### 3. VisibleWhen for conditional fields
**Choice**: 8 of 9 fields are conditionally visible based on `os_sandbox_enabled`.
**Rationale**: Follows the P2P sandbox form pattern where container-specific fields are gated behind `container_enabled`.
