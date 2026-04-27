## 1. Approval-Flow Domain Model

- [x] 1.1 Add the `internal/approvalflow` package with approval objects, decision states, issue classes, fulfillment grades, and release outcome records
- [x] 1.2 Add the core `ApproveArtifactRelease` evaluator with approve/reject/request-revision/escalate behavior

## 2. Artifact Release Approval Receipts

- [x] 2.1 Extend the audit log schema with `artifact_release_approval`
- [x] 2.2 Add the `approve_artifact_release` meta tool and audit-backed release receipts

## 3. Operator Surface And Docs

- [x] 3.1 Add minimal operator-facing approval-flow docs and truth-align surrounding surfaces

## 4. Verification And OpenSpec Closeout

- [x] 4.1 Run targeted tests while implementing each slice
- [x] 4.2 Run `go test ./...`, `go build ./...`, and `python3 -m mkdocs build --strict`
- [x] 4.3 Sync main specs and archive the completed change
