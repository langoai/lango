## Why

The `tool_output_get` surface already relies on required wrapper inputs: `ref` is always required, and `pattern` is effectively required in `grep` mode. But there was no direct tool-entrypoint regression proving those failures happen before stored-output lookup, and the docs only partially described the `grep`-mode requirement.

## What Changes

- Add direct tool-entrypoint regressions for missing `ref` and missing `pattern` in grep mode.
- Update README and TOOL_USAGE wording to describe the retrieval contract precisely.
- Sync proactive-output-gatekeeper, downstream-docs, and production-readiness specs with the same contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `proactive-output-gatekeeper`: `tool_output_get` required-input guards are now directly covered.
- `downstream-docs-sync`: output-retrieval docs now mention `ref` and grep-mode `pattern`.
- `production-readiness`: wrapper guard coverage now includes output retrieval.

## Impact

- Affected tests: `internal/tooloutput/tools_test.go`
- Affected prompts/docs: `prompts/TOOL_USAGE.md`, `README.md`
- Affected specs: `openspec/specs/proactive-output-gatekeeper/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`, `openspec/specs/production-readiness/spec.md`
