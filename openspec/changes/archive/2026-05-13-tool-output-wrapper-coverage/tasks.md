## 1. Guard Coverage

- [x] 1.1 Add tool-entrypoint regression coverage for missing `ref` on `tool_output_get`.
- [x] 1.2 Add tool-entrypoint regression coverage for missing grep `pattern`.

## 2. Prompt And Docs Sync

- [x] 2.1 Update TOOL_USAGE and README for the output retrieval input contract.
- [x] 2.2 Update proactive-output-gatekeeper, downstream-docs, and production-readiness specs for the same contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/tooloutput -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
