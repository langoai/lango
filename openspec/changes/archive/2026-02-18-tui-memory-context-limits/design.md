## Context

The Observational Memory form currently exposes 6 fields (enabled, provider, model, message threshold, observation threshold, max budget). Two new config fields (`MaxReflectionsInContext`, `MaxObservationsInContext`) need TUI support.

## Goals / Non-Goals

**Goals:**
- Expose `maxReflectionsInContext` and `maxObservationsInContext` in the settings form
- Wire state update handler for both fields

**Non-Goals:**
- No changes to config types or core logic (already implemented)
- No onboard wizard changes (onboard does not have an Observational Memory step)

## Decisions

1. **Validation**: Allow `>= 0` (non-negative) since `0` means unlimited, unlike the threshold fields which require positive values.
2. **Placement**: Add after the existing `om_max_budget` field to keep context-related settings grouped together.

## Risks / Trade-offs

- [Minimal risk] Thin UI-layer addition following established patterns.
