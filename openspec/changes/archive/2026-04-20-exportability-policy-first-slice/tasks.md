## 1. Core Policy Engine

- [x] 1.1 Add the `internal/exportability` package with source classes, decision states, receipt types, and a source-primary evaluator
- [x] 1.2 Add `security.exportability.enabled` config wiring and defaults for the first slice

## 2. Knowledge Asset Source Tagging

- [x] 2.1 Extend the `knowledge` Ent schema with `source_class` and `asset_label`
- [x] 2.2 Extend `knowledge.KnowledgeEntry` and store CRUD paths to persist and read source tagging metadata
- [x] 2.3 Treat exportability-class changes as version-significant in the knowledge store
- [x] 2.4 Extend `save_knowledge` tool parameters to accept source tagging metadata

## 3. Evaluation Receipts

- [x] 3.1 Extend the audit log schema with `exportability_decision`
- [x] 3.2 Add a batch knowledge lookup helper for evaluation inputs
- [x] 3.3 Add `evaluate_exportability` meta tool that evaluates source keys and writes audit-backed receipts

## 4. Operator Surface And Docs

- [x] 4.1 Surface exportability status in `lango security status`
- [x] 4.2 Add the security/operator docs for exportability and update architecture docs and README truthfully

## 5. Verification And OpenSpec Closeout

- [x] 5.1 Run targeted tests while implementing each slice
- [x] 5.2 Run `go test ./...`, `go build ./...`, and `python3 -m mkdocs build --strict`
- [x] 5.3 Verify the change against OpenSpec, sync main specs, and archive the completed change
