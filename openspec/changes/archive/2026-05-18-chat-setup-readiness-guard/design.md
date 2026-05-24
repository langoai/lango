# Design: Chat Setup Readiness Guard

## Current Behavior

Plain chat stores the loaded config in `ChatModel.cfg`, renders header state through `renderHeader`, renders turn affordance copy through `renderTurnStrip`, and submits non-slash input through `submitCurrentInput`.

The workbench already treats `!config.EvaluateAgentSetup(cfg).Ready()` as a first-run setup state. Plain chat does not use that shared readiness contract today, so the same default profile can be blocked in the workbench but treated as sendable in `lango chat`.

## Proposed Approach

### Readiness Source

Use `config.EvaluateAgentSetup(m.cfg)` directly from chat code. This keeps chat aligned with the existing config-level contract and avoids duplicating provider/model/API-key rules in the UI package.

### Rendering

Add setup-aware rendering helpers so the focused chat surface communicates the true state:

- Header: show a setup-required status instead of falling back to `default · auto` when readiness is incomplete.
- Turn strip: show setup-required label/copy while idle or failed and setup is incomplete.
- Help/footer affordances: replace the send-focused Enter hint with setup guidance while still keeping slash commands discoverable.

The implementation should preserve the current narrow-width truncation behavior.

### Submission Gate

In `submitCurrentInput`, keep slash-command dispatch first so inspection and help commands remain available. For normal input, check readiness before `onUserSubmission`, user transcript append, pending activation, state transition, or `TurnRunner.Run`.

When blocked, append a system/status message with actionable setup guidance:

- `lango onboard`
- `lango settings`
- `lango doctor`

The command should leave chat idle and should not clear the user's input unless the current input was a slash command. This avoids data loss and lets the user copy or revise what they wrote after setup.

### Ready Paths

Existing ready paths must continue to submit normally:

- Remote provider config with provider ID, model, provider type, and API key.
- Ollama provider config with provider ID, model, and provider type, without requiring an API key.

## Risks

- If slash-command dispatch happens after the readiness check, pre-setup diagnostics would be blocked. Tests should cover this ordering.
- If the input is cleared before readiness gating, users lose their draft. Tests should cover composer retention.
- If readiness rendering only changes the header, users may still see `Enter sends`; tests should cover turn strip/help copy.
