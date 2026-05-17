## 1. Planning

- [x] 1.1 Create focused OpenSpec artifacts for CLI stream stdin guard hardening
- [x] 1.2 Validate and commit planning artifacts

## 2. Tests First

- [ ] 2.1 Add a failing fixture regression proving direct `os.Stdin` is rejected

## 3. Implementation

- [ ] 3.1 Extract reusable CLI stream guard scanning helper
- [ ] 3.2 Reject direct `os.Stdin` references outside approved seam files
- [ ] 3.3 Preserve existing raw print and stdout/stderr checks

## 4. Review And Verification

- [ ] 4.1 Run focused internal/testutil stream guard tests
- [ ] 4.2 Complete local teammate Reviewer/QA review for guard scope and false positives
- [ ] 4.3 Run `go build ./...`
- [ ] 4.4 Run `go test ./...`
- [ ] 4.5 Run `git diff --check`
- [ ] 4.6 Sync main specs and run `openspec validate --all --strict`
- [ ] 4.7 Archive the completed OpenSpec change
