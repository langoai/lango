## 1. Planning

- [x] 1.1 Create focused OpenSpec artifacts for doctor network test stabilization
- [x] 1.2 Validate and commit planning artifacts

## 2. Implementation

- [ ] 2.1 Replace the fixed-port doctor availability test with an allocated loopback port
- [ ] 2.2 Assert the dynamic success message and listen address

## 3. Review And Verification

- [ ] 3.1 Run focused doctor checks tests
- [ ] 3.2 Complete local teammate Reviewer/QA review for test determinism
- [ ] 3.3 Run `go build ./...`
- [ ] 3.4 Run `go test ./...`
- [ ] 3.5 Run `git diff --check`
- [ ] 3.6 Sync main specs and run `openspec validate --all --strict`
- [ ] 3.7 Archive the completed OpenSpec change
