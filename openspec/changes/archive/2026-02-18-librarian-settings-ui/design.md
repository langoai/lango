## Context

The `LibrarianConfig` struct exists in `internal/config/types.go` with 7 fields but lacks default values in the config loader and has no Settings TUI integration. Every other subsystem (Cron, Background, Workflow, Payment) follows a consistent pattern: defaults in `DefaultConfig()`, viper `SetDefault` bindings in `Load()`, a menu entry, a form constructor, state update cases, and editor routing.

## Goals / Non-Goals

**Goals:**
- Add Librarian default values to `DefaultConfig()` and viper SetDefault bindings
- Expose all 7 Librarian fields in the Settings TUI with appropriate input types
- Follow existing patterns exactly for consistency

**Non-Goals:**
- Changing Librarian runtime behavior or its struct definition
- Adding validation logic beyond form-level input checks
- Modifying the proactive-librarian feature itself

## Decisions

1. **Field key prefix**: Use `lib_` prefix for all form field keys (e.g., `lib_enabled`, `lib_obs_threshold`) to avoid collisions with other config sections. This is consistent with how other sections use prefixes (`om_`, `wf_`, `bg_`, `cron_`).

2. **Menu placement**: Insert "Librarian" between "Workflow Engine" and "Save & Exit" to group it with other automation-adjacent features.

3. **AutoSaveConfidence input type**: Use `InputSelect` with options `["high", "medium", "low"]` rather than free text, since the field has a fixed set of valid values.

4. **Provider field**: Use `InputSelect` with `[""] + buildProviderOptions(cfg)` to allow empty (use agent default) or selecting a registered provider, matching the pattern used by Observational Memory.

## Risks / Trade-offs

- [Minimal risk] This is a purely additive UI change following established patterns. No existing behavior is modified.
