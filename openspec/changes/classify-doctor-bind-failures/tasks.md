## 1. Planning

- [x] 1.1 Create focused OpenSpec artifacts for doctor bind failure classification
- [x] 1.2 Validate and commit planning artifacts

## 2. Tests First

- [ ] 2.1 Add a failing doctor server port test for malformed bind-host diagnostics

## 3. Implementation

- [ ] 3.1 Classify address-in-use errors separately from other listen failures
- [ ] 3.2 Preserve existing occupied-port and IPv6 success behavior

## 4. Review And Verification

- [ ] 4.1 Run focused doctor network tests
- [ ] 4.2 Run subagent review for spec/test/diagnostic coverage
- [ ] 4.3 Run `go build ./...`
- [ ] 4.4 Run `go test ./...`
- [ ] 4.5 Run `git diff --check`
- [ ] 4.6 Sync main specs and run `openspec validate --all --strict`
- [ ] 4.7 Archive the completed OpenSpec change
