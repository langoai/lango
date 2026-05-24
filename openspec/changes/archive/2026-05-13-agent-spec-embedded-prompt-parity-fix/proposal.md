## Why

The initial built-in teammate identity source sync updated the embed surface and spec, but the runtime `agentSpecs[].Instruction` strings still diverged from several embedded `prompts/agents/<name>/IDENTITY.md` files. That left the new parity regression failing and meant the runtime could still present different built-in instructions than the embedded defaults.

## What Changes

- Make `agentSpecs[].Instruction` consistently prefer embedded per-agent identity files.
- Align the orchestration parity tests with the current built-in escalation wording and whitespace/format normalization expectations.
- Verify that runtime agent specs and embedded per-agent identities now stay in lockstep.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `sub-agent-default-prompts`: runtime built-in teammate instructions now stay aligned with embedded per-agent identity prompts.

## Impact

- Affected code: `internal/orchestration/tools.go`, `internal/orchestration/orchestrator_test.go`, `prompts/embed.go`
- Verification: `go test ./internal/orchestration -count=1`, `go build ./...`, `go test ./...`
