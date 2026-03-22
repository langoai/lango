## Context

`config.Provenance` already exists with stable config fields:

- `enabled`
- `checkpoints.autoOnStepComplete`
- `checkpoints.autoOnPolicy`
- `checkpoints.maxPerSession`
- `checkpoints.retentionDays`

The settings editor already groups adjacent automation/runtime controls in one section:

- Cron
- Background
- Workflow
- RunLedger

Provenance belongs next to those controls because it is config-backed, automation-adjacent, and already part of the application configuration model.

`session_isolation` is intentionally not part of this change because it is not config-backed. It lives in agent definitions (`AGENT.md`) and needs an agent-definition editing UX, not a global settings form.

## Goals / Non-Goals

**Goals**
- Add a provenance settings form for the existing config fields only
- Keep placement in the Automation section next to RunLedger
- Preserve current settings editor patterns and validation behavior

**Non-Goals**
- Adding new provenance config fields
- Exposing CLI one-off flags such as bundle redaction in settings
- Adding AGENT.md editing or `session_isolation` controls

## Decisions

### D1. Provenance lives in Automation

The category is added to the Automation section after RunLedger because the settings editor already groups durable execution/runtime control surfaces there.

### D2. Use existing config shape verbatim

The form maps 1:1 to the existing `config.Provenance` fields. No new config or compatibility layer is introduced.

### D3. Validation mirrors existing form conventions

Boolean fields use toggles. Integer fields (`maxPerSession`, `retentionDays`) require `0` or greater, matching the existing meaning of `0 = unlimited`.

### D4. session_isolation stays out

`session_isolation` remains AGENT.md metadata and is documented as such. This change does not try to overload the settings editor with agent-definition concerns.
