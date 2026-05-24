## Why

`knowledge exchange v1` needs an explicit answer to whether a deliverable may cross the local trust boundary. Today Lango has privacy, approval, provenance, and audit subsystems, but it still lacks a first-class exportability policy with durable decision records.

## What Changes

- Introduce a source-primary exportability evaluator for early knowledge exchange artifacts.
- Persist source classification metadata on knowledge assets so the evaluator has explicit lineage input.
- Add an exportability evaluation tool that emits durable `exportability_decision` receipts into audit storage.
- Surface the first-slice policy status in security/operator docs and CLI status output.

## Capabilities

### New Capabilities
- `exportability-policy`: Source-primary artifact exportability evaluation with `exportable`, `blocked`, and `needs-human-review` states plus receipt-style decision records.

### Modified Capabilities
- `knowledge-store`: Knowledge assets gain source-class and asset-label metadata used by exportability evaluation.
- `meta-tools`: Meta tools add exportability-aware source tagging on `save_knowledge` and a new `evaluate_exportability` tool.
- `cli-security-status`: Security status surfaces exportability-policy state.
- `security-docs-sync`: Security and architecture docs describe the new exportability-policy operator model truthfully.

## Impact

- Affected code: `internal/exportability/*`, `internal/knowledge/*`, `internal/app/tools_meta.go`, `internal/cli/security/status.go`, `internal/config/*`
- Affected storage: Ent schema for `knowledge` and `audit_log`
- Affected operator surface: `lango security status`, security docs, architecture docs, README
