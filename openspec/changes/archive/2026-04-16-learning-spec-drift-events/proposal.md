## Why

The learning engine's `SuggestionEmitter` surfaces tool-level patterns (new rules/skills) but has no mechanism to signal "spec drift" — when accumulated learning patterns indicate the OpenSpec documentation may be stale. The Managed Agents article emphasizes compound learning (AGENTS.md / spec artifacts auto-updating). Rather than having the runtime write to `openspec/changes/` directly (high noise risk), this change adds a `SpecDriftDetectedEvent` that signals the drift without generating artifacts. Operators can then manually update specs based on the signal.

## What Changes

- Add `SpecDriftDetectedEvent` to `eventbus/continuity_events.go`
- Add `EmitSpecDrift` method to `SuggestionEmitter` that evaluates accumulated patterns for spec drift indicators and publishes the event
- The spec drift signal fires when the same tool error pattern recurs above a frequency threshold across sessions, suggesting the spec's expected behavior no longer matches reality
- No automatic OpenSpec draft creation — event only

## Capabilities

### New Capabilities
- `learning-spec-drift`: Spec drift detection event from learning engine

### Modified Capabilities
_(none)_

## Impact

- `internal/eventbus/continuity_events.go` — new event type
- `internal/learning/suggestion.go` — new `EmitSpecDrift` method + drift tracking state
- `internal/learning/engine.go` — wire drift check into existing `OnToolResult` observer path
