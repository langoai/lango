## Context

Bare `lango` is the product's primary interactive entry point. The current empty Mission Control screen is generic and assumes the operator can immediately chat, but the default auto-created profile still lacks a usable provider map and model setup. The result is an attractive but misleading empty state.

## Goals / Non-Goals

**Goals:**
- Detect the common incomplete-profile case directly in the workbench Mission Control page.
- Present clear next steps for setup, full editing, and verification.
- Avoid showing this setup guidance once the profile is plausibly ready for normal use.
- Keep cockpit-only surfaces unchanged except for shared page plumbing required by the workbench.

**Non-Goals:**
- Changing runtime configuration semantics.
- Expanding onboarding scope.
- Redesigning the overall Mission Control layout or composer UX.

## Decisions

### D1: Use lightweight readiness heuristics in the page layer
The workbench page only needs to detect obviously incomplete profiles. The chosen heuristic is intentionally simple: missing provider ID, missing model, missing provider entry, missing provider type, or missing API key for non-Ollama providers.

### D2: Limit setup guidance to the workbench surface
The explicit cockpit is an advanced surface; the dead-end problem is most acute on bare `lango`. The setup hint therefore appears only on the workbench-flavored Mission Control surface.

### D3: Lock the behavior with workbench-level regression tests
Tests assert both sides of the behavior: incomplete config shows `onboard/settings/doctor`, while a ready config omits those setup prompts.

## Risks / Trade-offs

- **[Heuristic is conservative]** → Some configs may still be technically unusable even if the hint disappears. Mitigation: the guidance is meant to catch the obvious first-run case, not replace deeper runtime validation.
- **[More text in empty state]** → The empty state gets slightly longer. Mitigation: the added copy is only shown when the profile is incomplete and directly improves operator recovery.
