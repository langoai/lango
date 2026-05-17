## 1. Planning

- [x] 1.1 Create OpenSpec artifacts for gateway target normalization
- [x] 1.2 Validate and commit planning artifacts

## 2. Tests First

- [x] 2.1 Add failing clihttp resolver test for explicit trailing-slash normalization
- [x] 2.2 Add failing status command test proving `--addr` display matches normalized probe target

## 3. Implementation

- [x] 3.1 Normalize explicit gateway addresses in the shared resolver
- [x] 3.2 Make root status output report the resolved explicit gateway target
- [x] 3.3 Preserve existing config-derived status behavior for direct collection callers

## 4. Review And Verification

- [x] 4.1 Run focused package tests
- [x] 4.2 Run subagent review for the target-normalization change
- [x] 4.3 Run `go build ./...`
- [x] 4.4 Run `go test ./...`
- [x] 4.5 Run `git diff --check`
- [x] 4.6 Sync main specs and run `openspec validate --all --strict`
- [x] 4.7 Archive the completed OpenSpec change
