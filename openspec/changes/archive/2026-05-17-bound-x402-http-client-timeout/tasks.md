## 1. Planning

- [x] 1.1 Create focused OpenSpec artifacts for bounded X402 HTTP client timeout
- [x] 1.2 Validate and commit planning artifacts

## 2. Regression Coverage

- [x] 2.1 Add a regression proving `Interceptor.HTTPClient` returns a client with a finite timeout
- [x] 2.2 Add coverage that cached client reuse preserves the bounded client

## 3. Implementation

- [x] 3.1 Add a finite default timeout for the X402 wrapped base HTTP client
- [x] 3.2 Wire the bounded base client into the Coinbase SDK wrapper without changing spending-limit behavior

## 4. Review And Verification

- [x] 4.1 Run focused X402 tests
- [x] 4.2 Run subagent-driven spec and code-quality review checkpoints
- [x] 4.3 Run `go build ./...`
- [x] 4.4 Run `go test ./...`
- [x] 4.5 Run `git diff --check`
- [x] 4.6 Sync main specs and run `openspec validate --all --strict`
- [x] 4.7 Archive the completed OpenSpec change
