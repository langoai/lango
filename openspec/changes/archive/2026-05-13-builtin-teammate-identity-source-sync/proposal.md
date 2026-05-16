## Why

The embedded `prompts/agents/<name>/IDENTITY.md` files and the inline `agentSpecs[].Instruction` strings had started to drift apart. That left the runtime prompt surface and the embedded default prompt surface with two different sources of truth.

## What Changes

- Extend `prompts.FS` to embed agent prompt files.
- Make `agentSpecs[].Instruction` prefer the embedded `prompts/agents/<name>/IDENTITY.md` content, falling back only if the embedded read fails.
- Add a regression that checks embedded prompt content against the runtime agent spec instructions.
- Update sub-agent-default-prompts spec to describe the embedded file as the preferred source for built-in teammate instructions.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `sub-agent-default-prompts`: built-in teammate instructions now use the embedded per-agent prompt files as their preferred source of truth.

## Impact

- Affected code: `prompts/embed.go`, `internal/orchestration/tools.go`, `internal/orchestration/orchestrator_test.go`
- Affected specs: `openspec/specs/sub-agent-default-prompts/spec.md`
