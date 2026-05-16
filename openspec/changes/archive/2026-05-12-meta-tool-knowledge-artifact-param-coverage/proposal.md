## Why

The transaction and receipts-oriented meta tool wrappers are now heavily covered, but several foundational knowledge-artifact tools still rely on implicit `toolparam` behavior for their required inputs. Those tools are still operator-facing entrypoints and should preserve explicit wrapper-level missing-parameter contracts too.

## What Changes

- Add wrapper-level missing-parameter regressions for `save_knowledge`, `evaluate_exportability`, and `approve_artifact_release`.
- Extend meta-tools and production-readiness coverage for those knowledge-artifact wrapper guards.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `meta-tools`: the foundational knowledge-artifact tools now have explicit wrapper-level missing-parameter coverage.
- `production-readiness`: wrapper-level request-guard coverage now includes the core knowledge save/exportability/release-approval tool cluster.

## Impact

- Affected tests: `internal/app/tools_meta_validation_test.go`, `internal/app/tools_meta_exportability_test.go`, `internal/app/tools_meta_approvalflow_test.go`
- Affected specs: `openspec/specs/meta-tools/spec.md`, `openspec/specs/production-readiness/spec.md`
