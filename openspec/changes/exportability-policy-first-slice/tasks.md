## 1. Core Policy Engine

- [x] 1.1 Add the `internal/exportability` package with source classes, decision states, receipt types, and a source-primary evaluator
- [x] 1.2 Add `security.exportability.enabled` config wiring and defaults for the first slice

## 2. Knowledge Asset Source Tagging

- [ ] 2.1 Extend the `knowledge` Ent schema with `source_class` and `asset_label`
- [ ] 2.2 Extend `knowledge.KnowledgeEntry` and store CRUD paths to persist and read source tagging metadata
- [ ] 2.3 Treat exportability-class changes as version-significant in the knowledge store
- [ ] 2.4 Extend `save_knowledge` tool parameters to accept source tagging metadata

## 3. Evaluation Receipts

- [ ] 3.1 Extend the audit log schema with `exportability_decision`
- [ ] 3.2 Add a batch knowledge lookup helper for evaluation inputs
- [ ] 3.3 Add `evaluate_exportability` meta tool that evaluates source keys and writes audit-backed receipts

## 4. Operator Surface And Docs

- [ ] 4.1 Surface exportability status in `lango security status`
- [ ] 4.2 Add the security/operator docs for exportability and update architecture docs and README truthfully

## 5. Verification And OpenSpec Closeout

- [ ] 5.1 Run targeted tests while implementing each slice
- [ ] 5.2 Run `go test ./...`, `go build ./...`, and `python3 -m mkdocs build --strict`
- [ ] 5.3 Verify the change against OpenSpec, sync main specs, and archive the completed change
