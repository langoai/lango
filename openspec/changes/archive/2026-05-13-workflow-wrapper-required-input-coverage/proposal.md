## Why

The workflow tool cluster already declares `run_id`, `name`, and `yaml_content` as required wrapper inputs where applicable, and the handlers reject missing values. But there was no direct regression proving those failures happen before workflow lookup or file writes, and the automator prompt plus user-facing docs did not state that contract explicitly.

## What Changes

- Add direct tool-entrypoint regressions for missing `run_id`, `name`, and `yaml_content` across `workflow_*` tools.
- Update automator runtime and embedded prompts to describe the workflow required-input contract.
- Sync TOOL_USAGE, README, multi-agent docs, and workflow/production specs with the same contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `automation-agent-tools`: `workflow_*` required-input guards are now directly covered.
- `multi-agent-orchestration`: automator prompt wording now explicitly covers the `workflow_*` input contract.
- `downstream-docs-sync`: workflow docs now mention the required wrapper inputs.
- `production-readiness`: wrapper guard coverage now includes workflow tools.

## Impact

- Affected code: `internal/orchestration/tools.go`
- Affected tests: `internal/workflow/tools_test.go`
- Affected prompts: `prompts/TOOL_USAGE.md`, `prompts/agents/automator/IDENTITY.md`
- Affected docs: `README.md`, `docs/features/multi-agent.md`
- Affected specs: `openspec/specs/automation-agent-tools/spec.md`, `openspec/specs/agent-prompting/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`, `openspec/specs/production-readiness/spec.md`
