## 1. Planning

- [x] 1.1 Create focused OpenSpec artifacts for gateway shutdown lifecycle hardening
- [x] 1.2 Validate and commit planning artifacts

## 2. Regression Coverage

- [ ] 2.1 Add a regression for `Shutdown()` before `Start()`
- [ ] 2.2 Add a regression for `Shutdown()` after a failed `Start()`
- [ ] 2.3 Add a regression proving repeated shutdown calls do not panic

## 3. Implementation

- [ ] 3.1 Make gateway shutdown nil-safe for pre-start and failed-start states
- [ ] 3.2 Preserve graceful shutdown behavior for running servers

## 4. Review And Verification

- [ ] 4.1 Run focused gateway tests
- [ ] 4.2 Run subagent-driven spec and code-quality review checkpoints
- [ ] 4.3 Run `go build ./...`
- [ ] 4.4 Run `go test ./...`
- [ ] 4.5 Run `git diff --check`
- [ ] 4.6 Sync main specs and run `openspec validate --all --strict`
- [ ] 4.7 Archive the completed OpenSpec change
