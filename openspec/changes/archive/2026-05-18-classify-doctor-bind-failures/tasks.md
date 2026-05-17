## 1. Planning

- [x] 1.1 Create focused OpenSpec artifacts for doctor bind failure classification
- [x] 1.2 Validate and commit planning artifacts

## 2. Tests First

- [x] 2.1 Add a failing doctor server port test for malformed bind-host diagnostics

## 3. Implementation

- [x] 3.1 Classify address-in-use errors separately from other listen failures
- [x] 3.2 Preserve existing occupied-port and IPv6 success behavior

## 4. Review And Verification

- [x] 4.1 Run focused doctor network tests
- [x] 4.2 Attempt subagent review for spec/test/diagnostic coverage; usage limit blocked fresh review, so complete local teammate Reviewer/QA review
- [x] 4.3 Run `go build ./...`
- [x] 4.4 Run `go test ./...`
- [x] 4.5 Run `git diff --check`
- [x] 4.6 Sync main specs and run `openspec validate --all --strict`
- [x] 4.7 Archive the completed OpenSpec change
