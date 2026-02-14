## 1. Dead Code Removal

- [x] 1.1 Delete `UpdateField` method from `internal/cli/onboard/state.go` (lines 36-41)

## 2. Comment Clarification

- [x] 2.1 Replace `// TODO: Implement save logic` in `internal/cli/onboard/wizard.go:212` with descriptive comment
- [x] 2.2 Replace `// TODO: Implement ListModels proxying if needed` in `internal/supervisor/proxy.go:111-113` with descriptive comment
- [x] 2.3 Update corresponding test comment in `internal/supervisor/supervisor_test.go:541`

## 3. Timeout Constant Extraction

- [x] 3.1 Add `rpcTimeout = 30 * time.Second` package-level constant to `internal/security/rpc_provider.go`
- [x] 3.2 Replace 3 hardcoded `time.After(30 * time.Second)` calls with `time.After(rpcTimeout)` and remove `// TODO: Configurable timeout?`

## 4. Verification

- [x] 4.1 Run `go build ./...` — no errors
- [x] 4.2 Run `go test ./...` — no regressions from this change
