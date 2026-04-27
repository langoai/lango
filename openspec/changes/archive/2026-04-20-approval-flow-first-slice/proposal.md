## Why

`knowledge exchange v1` now has a first exportability slice, but it still lacks a product-level approval flow for artifact release. Without that, Lango cannot clearly decide when a leader agent may release, reject, revise, or escalate a deliverable, or how that decision should shape settlement.

## What Changes

- Introduce a first-slice approval-flow domain model centered on `artifact release approval`.
- Add structured approval decisions and outcome records that later settlement and dispute systems can consume.
- Add an agent-facing artifact-release approval tool that writes audit-backed approval receipts.
- Add a minimal operator-facing approval-flow doc and truth-align the surrounding architecture/security docs.

## Capabilities

### New Capabilities
- `approval-flow`: Structured approval objects, decision states, and release outcome records for `knowledge exchange v1`.

### Modified Capabilities
- `meta-tools`: Add an `approve_artifact_release` tool and related audit-backed receipt behavior.
- `security-docs-sync`: Add truthful approval-flow operator documentation and link it from the security/architecture surfaces.

## Impact

- Affected code: `internal/approvalflow/*`, `internal/app/tools_meta.go`, `internal/ent/schema/audit_log.go`
- Affected storage: audit log action enum and receipt persistence path
- Affected docs: security docs, architecture track/audit docs, README
