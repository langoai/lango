## 1. Tests
- [x] 1.1 Add failing sandbox status tests for JSON output, plain output, invalid output pre-load rejection, and default table preservation.

## 2. Implementation
- [x] 2.1 Refactor sandbox status data collection into a snapshot that preserves existing graceful bootstrap/config fallback behavior.
- [x] 2.2 Add `--output table|json|plain` rendering for `lango sandbox status`.
- [x] 2.3 Keep current default table output behavior compatible with existing tests.

## 3. Specs, Docs, and Verification
- [x] 3.1 Update the main `os-sandbox-cli` spec with the output-format contract.
- [x] 3.2 Update public sandbox CLI docs to document the new `--output` flag and JSON fields.
- [x] 3.3 Validate the OpenSpec change in strict mode.
- [x] 3.4 Run focused sandbox CLI tests.
- [x] 3.5 Run `go build ./...` and `go test ./...`.
- [x] 3.6 Run subagent-driven review.
- [x] 3.7 Archive the OpenSpec change and commit this scoped unit separately.
