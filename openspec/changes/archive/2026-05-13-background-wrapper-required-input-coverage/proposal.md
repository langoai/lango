## Why

The background tool cluster already declares `prompt` and `task_id` as required wrapper inputs, and the handlers reject missing values. But there was no direct regression proving those failures happen before queue submission or task lookup, and the automator prompt plus user-facing docs did not state that contract explicitly.

## What Changes

- Add direct tool-entrypoint regressions for missing `prompt` and `task_id` across `bg_*` tools.
- Update automator runtime and embedded prompts to describe the required-input contract.
- Sync TOOL_USAGE, README, multi-agent docs, and background/production specs with the same contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `background-execution`: `bg_*` required-input guards are now directly covered.
- `multi-agent-orchestration`: automator prompt wording now explicitly covers the `bg_*` input contract.
- `downstream-docs-sync`: background-task docs now mention the required wrapper inputs.
- `production-readiness`: wrapper guard coverage now includes background-task tools.

## Impact

- Affected code: `internal/orchestration/tools.go`
- Affected tests: `internal/background/tools_test.go`
- Affected prompts: `prompts/TOOL_USAGE.md`, `prompts/agents/automator/IDENTITY.md`
- Affected docs: `README.md`, `docs/features/multi-agent.md`
- Affected specs: `openspec/specs/background-execution/spec.md`, `openspec/specs/agent-prompting/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`, `openspec/specs/production-readiness/spec.md`
