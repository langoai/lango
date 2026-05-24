## Why

The vault security tools already declare required inputs such as `data`, `ciphertext`, `name`, and `value`, and their handlers reject missing values. But the agent-facing tool entrypoints did not have direct regression coverage for those missing-input paths, and the prompt plus operator-facing docs did not state the required-input contract explicitly.

## What Changes

- Add tool-entrypoint regressions for missing required inputs across crypto and secrets tools.
- Update the shared tool-usage prompt to describe the vault security required-input contract.
- Sync README, CLI docs, and security/production/downstream specs with the same contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `security-tools`: crypto and secrets required-input guards are now directly covered at the tool entrypoint.
- `downstream-docs-sync`: vault security docs now mention the required-input contract.
- `production-readiness`: wrapper/input guard coverage now includes vault crypto and secrets tools.

## Impact

- Affected tests: `internal/tools/crypto/tools_test.go`, `internal/tools/secrets/tools_test.go`
- Affected prompts: `prompts/TOOL_USAGE.md`
- Affected docs: `README.md`, `docs/cli/agent-memory.md`
- Affected specs: `openspec/specs/security-tools/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`, `openspec/specs/production-readiness/spec.md`
