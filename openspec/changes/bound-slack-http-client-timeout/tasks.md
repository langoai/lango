## 1. Planning

- [x] 1.1 Create focused OpenSpec artifacts for Slack HTTP timeout hardening
- [x] 1.2 Validate and commit planning artifacts

## 2. Regression Coverage

- [ ] 2.1 Add a regression proving default Slack HTTP clients have a finite timeout
- [ ] 2.2 Add a regression proving injected Slack HTTP clients are preserved

## 3. Implementation

- [ ] 3.1 Add a finite default Slack HTTP client timeout
- [ ] 3.2 Wire the bounded default into Slack SDK client construction

## 4. Review And Verification

- [ ] 4.1 Run focused Slack channel tests
- [ ] 4.2 Run subagent-driven spec and code-quality review checkpoints
- [ ] 4.3 Run `go build ./...`
- [ ] 4.4 Run `go test ./...`
- [ ] 4.5 Run `git diff --check`
- [ ] 4.6 Sync main specs and run `openspec validate --all --strict`
- [ ] 4.7 Archive the completed OpenSpec change
