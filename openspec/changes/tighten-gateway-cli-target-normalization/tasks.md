## 1. Planning

- [x] 1.1 Create OpenSpec artifacts for gateway target normalization
- [ ] 1.2 Validate and commit planning artifacts

## 2. Tests First

- [ ] 2.1 Add failing clihttp resolver test for explicit trailing-slash normalization
- [ ] 2.2 Add failing status command test proving `--addr` display matches normalized probe target

## 3. Implementation

- [ ] 3.1 Normalize explicit gateway addresses in the shared resolver
- [ ] 3.2 Make root status output report the resolved explicit gateway target
- [ ] 3.3 Preserve existing config-derived status behavior for direct collection callers

## 4. Review And Verification

- [ ] 4.1 Run focused package tests
- [ ] 4.2 Run subagent review for the target-normalization change
- [ ] 4.3 Run `go build ./...`
- [ ] 4.4 Run `go test ./...`
- [ ] 4.5 Run `git diff --check`
- [ ] 4.6 Sync main specs and run `openspec validate --all --strict`
- [ ] 4.7 Archive the completed OpenSpec change
