## Overview

This change introduces a single `config.EvaluateAgentSetup` helper that returns a structured readiness snapshot for the active agent profile.

## Design Decisions

### Shared readiness snapshot

The helper returns normalized fields for provider ID, model, provider type, and explicit failure modes:

- missing provider
- missing model
- missing provider mapping
- missing provider type
- missing API key

This keeps UI and validation surfaces from re-deriving the same conditions with slightly different behavior.

### Ollama exception stays explicit

Ollama remains the only built-in provider path that does not require an API key for ready-state UX and onboarding validation. The helper preserves that rule in one place.

### Nil config remains non-ready for setup checks

General readiness evaluation treats a nil config as incomplete. The Mission Control projector still keeps its local nil-config fast path to avoid changing unrelated empty-header behavior outside the workbench-configured flow.
