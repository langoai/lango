## Why

The first wrapper-level regression pass covered the settlement and escrow-execution write tools, but several adjacent transaction-receipt-backed meta tools still relied on implicit `toolparam` behavior without direct tests. Those tools are still operator-facing entrypoints, and the missing-parameter contract should be pinned the same way.

## What Changes

- Add wrapper-level missing-parameter regressions for dispute-hold, escrow release, escrow refund, post-adjudication status, and post-adjudication replay meta tools.
- Extend meta-tools and production-readiness coverage to require actionable wrapper-level parameter errors for those adjacent tools too.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `meta-tools`: additional transaction-receipt-backed operator tools now have explicit wrapper-level missing-parameter coverage.
- `production-readiness`: wrapper-level request guards stay explicit across the broader transaction-receipt-backed tool cluster.

## Impact

- Affected tests: `internal/app/tools_meta_disputehold_test.go`, `internal/app/tools_meta_escrowrelease_test.go`, `internal/app/tools_meta_escrowrefund_test.go`, `internal/app/tools_meta_postadjudicationstatus_test.go`, `internal/app/tools_meta_postadjudicationreplay_test.go`
- Affected specs: `openspec/specs/meta-tools/spec.md`, `openspec/specs/production-readiness/spec.md`
