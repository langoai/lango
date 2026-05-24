## Why

The cron tool cluster already declares `name`, `schedule_type`, `schedule`, `prompt`, and `id` as required wrapper inputs where applicable, and the handlers reject missing values. But there was no direct regression proving those failures happen before scheduler lookup or mutation, and the automator prompt plus user-facing docs did not state that contract explicitly.

## What Changes

- Add direct tool-entrypoint regressions for missing required inputs across `cron_*` tools.
- Update automator runtime and embedded prompts to describe the cron required-input contract.
- Sync TOOL_USAGE, README, multi-agent docs, and cron/production specs with the same contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `cron-scheduling`: `cron_*` required-input guards are now directly covered.
- `multi-agent-orchestration`: automator prompt wording now explicitly covers the `cron_*` input contract.
- `downstream-docs-sync`: cron docs now mention the required wrapper inputs.
- `production-readiness`: wrapper guard coverage now includes cron tools.

## Impact

- Affected tests: `internal/cron/tools_test.go`
- Affected prompts: `prompts/TOOL_USAGE.md`, `prompts/agents/automator/IDENTITY.md`
- Affected docs: `README.md`, `docs/features/multi-agent.md`
- Affected specs: `openspec/specs/cron-scheduling/spec.md`, `openspec/specs/agent-prompting/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`, `openspec/specs/production-readiness/spec.md`
