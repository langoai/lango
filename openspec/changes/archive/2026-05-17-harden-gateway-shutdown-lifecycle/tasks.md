## 1. Planning

- [x] 1.1 Create focused OpenSpec artifacts for gateway shutdown lifecycle hardening
- [x] 1.2 Validate and commit planning artifacts

## 2. Regression Coverage

- [x] 2.1 Add a regression for `Shutdown()` before `Start()`
- [x] 2.2 Add a regression for `Shutdown()` after a failed `Start()`
- [x] 2.3 Add a regression proving repeated shutdown calls do not panic

## 3. Implementation

- [x] 3.1 Make gateway shutdown nil-safe for pre-start and failed-start states
- [x] 3.2 Preserve graceful shutdown behavior for running servers

## 4. Review And Verification

- [x] 4.1 Run focused gateway tests
- [x] 4.2 Run subagent-driven spec and code-quality review checkpoints
- [x] 4.3 Run `go build ./...`
- [x] 4.4 Run `go test ./...`
- [x] 4.5 Run `git diff --check`
- [x] 4.6 Sync main specs and run `openspec validate --all --strict`
- [x] 4.7 Archive the completed OpenSpec change
