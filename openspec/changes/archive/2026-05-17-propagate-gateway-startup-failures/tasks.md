## 1. Planning

- [x] 1.1 Create focused OpenSpec artifacts for gateway startup failure propagation
- [x] 1.2 Validate and commit planning artifacts

## 2. Regression Coverage

- [x] 2.1 Add an application lifecycle regression for occupied gateway bind addresses
- [x] 2.2 Add a serve command regression proving startup summaries are suppressed when app startup fails

## 3. Implementation

- [x] 3.1 Split gateway binding from serving while preserving `Start()` behavior
- [x] 3.2 Update application lifecycle startup to propagate synchronous gateway listen failures

## 4. Review And Verification

- [x] 4.1 Run focused gateway/app/CLI tests
- [x] 4.2 Run subagent-driven spec and code-quality review checkpoints
- [x] 4.3 Run `go build ./...`
- [x] 4.4 Run `go test ./...`
- [x] 4.5 Run `git diff --check`
- [x] 4.6 Sync main specs and run `openspec validate --all --strict`
- [x] 4.7 Archive the completed OpenSpec change
