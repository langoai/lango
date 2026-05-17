## 1. Planning

- [x] 1.1 Create focused OpenSpec artifacts for CLI stream stdin guard hardening
- [x] 1.2 Validate and commit planning artifacts

## 2. Tests First

- [x] 2.1 Add a failing fixture regression proving direct `os.Stdin` is rejected

## 3. Implementation

- [x] 3.1 Extract reusable CLI stream guard scanning helper
- [x] 3.2 Reject direct `os.Stdin` references outside approved seam files
- [x] 3.3 Preserve existing raw print and stdout/stderr checks

## 4. Review And Verification

- [x] 4.1 Run focused internal/testutil stream guard tests
- [x] 4.2 Complete local teammate Reviewer/QA review for guard scope and false positives
- [x] 4.3 Run `go build ./...`
- [x] 4.4 Run `go test ./...`
- [x] 4.5 Run `git diff --check`
- [x] 4.6 Sync main specs and run `openspec validate --all --strict`
- [x] 4.7 Archive the completed OpenSpec change
