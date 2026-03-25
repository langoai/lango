## Context

The TUI settings editor uses a form-based system: `FormModel` with `Field` structs → `UpdateConfigFromForm` switch/case → `Config` struct. Orchestration config (`OrchestrationConfig` in `types_orchestration.go`) and trace store config (`TraceStoreConfig` in `types_observability.go`) already exist with default value functions, but have no form fields or state bindings.

The existing Multi-Agent form has 5 fields (`multi_agent`, `max_delegation_rounds`, `max_turns`, `error_correction_enabled`, `agents_dir`). The Observability form has 10 fields with `obs_` prefix. Both use `VisibleWhen` closures for conditional visibility.

## Goals / Non-Goals

**Goals:**
- Expose orchestration policy settings (circuit breaker, budget, recovery) in the existing Multi-Agent form
- Expose trace store retention settings in the existing Observability form
- Maintain backward compatibility with existing profiles
- Use `VisibleWhen` chain so structured-only fields appear only when relevant

**Non-Goals:**
- Creating new menu categories or form dispatchers
- Modifying config struct definitions or default value functions
- Adding CLI `config set` support for these fields (future work)
- Runtime validation beyond what the form validators enforce

## Decisions

### 1. Extend existing forms instead of creating new categories
Orchestration fields go into `NewMultiAgentForm`, trace store fields into `NewObservabilityForm`. This keeps the menu structure unchanged and avoids touching `menu.go` or `setup_flow.go`.

**Alternative considered:** Separate "Orchestration" category → rejected because it would fragment related settings and require dispatcher/menu changes.

### 2. Two-level VisibleWhen chain for orchestration
- Level 1: `orchestration_mode` visible when `multiAgentField.Checked`
- Level 2: All `orc_*` fields visible when `multiAgentField.Checked && modeField.Value == "structured"`

This prevents showing advanced policy fields when multi-agent is off or mode is classic.

### 3. Zero-value = runtime default
Runtime code (`delegation_guard.go`, `budget.go`, `recovery.go`) treats `<= 0` as "use default". UI validation allows `>= 0` and documents `(0 = use default: N)` in descriptions. This keeps the UI-runtime contract simple.

**Alternative considered:** Stricter validation (`> 0` only) → rejected because it would create a UI/runtime contract mismatch where UI rejects values the runtime handles safely.

### 4. Named field variables for VisibleWhen closures
The `multi_agent` and `obs_enabled` fields are refactored from anonymous `&tuicore.Field{}` to named variables (`multiAgentField`, `obsEnabledField`) so child closures can reference their state.

## Risks / Trade-offs

- [Hidden but preserved values] When `multi_agent` is toggled off, orchestration field values are hidden but preserved in config. Users may not realize settings persist when invisible → **Mitigation:** This matches existing behavior across the settings system (e.g., telegram fields when telegram is disabled).
- [Default fallback in forms] `DefaultConfig()` does not initialize `Orchestration` or `TraceStore` fields. Forms must use `OrchestrationDefaults()`/`TraceStoreDefaults()` to show sensible values for zero-value configs → **Mitigation:** Per-field zero-check with fallback to defaults function.
