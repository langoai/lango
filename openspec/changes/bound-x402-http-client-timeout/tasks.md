## 1. Planning

- [x] 1.1 Create focused OpenSpec artifacts for bounded X402 HTTP client timeout
- [x] 1.2 Validate and commit planning artifacts

## 2. Regression Coverage

- [ ] 2.1 Add a regression proving `Interceptor.HTTPClient` returns a client with a finite timeout
- [ ] 2.2 Add coverage that cached client reuse preserves the bounded client

## 3. Implementation

- [ ] 3.1 Add a finite default timeout for the X402 wrapped base HTTP client
- [ ] 3.2 Wire the bounded base client into the Coinbase SDK wrapper without changing spending-limit behavior

## 4. Review And Verification

- [ ] 4.1 Run focused X402 tests
- [ ] 4.2 Run subagent-driven spec and code-quality review checkpoints
- [ ] 4.3 Run `go build ./...`
- [ ] 4.4 Run `go test ./...`
- [ ] 4.5 Run `git diff --check`
- [ ] 4.6 Sync main specs and run `openspec validate --all --strict`
- [ ] 4.7 Archive the completed OpenSpec change
