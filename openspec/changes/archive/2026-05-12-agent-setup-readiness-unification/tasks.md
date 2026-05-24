## 1. Shared Readiness Helper

- [x] 1.1 Add a reusable agent readiness evaluator under `internal/config`.
- [x] 1.2 Cover remote-provider API-key requirements and the Ollama exception with unit tests.

## 2. Consumer Alignment

- [x] 2.1 Update onboard Test Configuration checks to use the shared readiness evaluator.
- [x] 2.2 Update Mission Control/workbench readiness signals to use the shared readiness evaluator.

## 3. Verification

- [x] 3.1 Run `gofmt -w` on modified Go files.
- [x] 3.2 Run focused tests for `internal/config`, `internal/cli/onboard`, `internal/cli/cockpit`, and `internal/cli/workbench`.
- [x] 3.3 Run `go build ./...`.
- [x] 3.4 Run `go test ./...`.
- [ ] 3.5 Validate and archive the OpenSpec change.
