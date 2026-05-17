## 1. Planning

- [x] 1.1 Create focused OpenSpec artifacts for doctor port conflict wording
- [x] 1.2 Validate and commit planning artifacts

## 2. Tests First

- [ ] 2.1 Add failing doctor server port test for occupied-port message

## 3. Implementation

- [ ] 3.1 Change occupied-port failure message to `Port <port> in use`

## 4. Review And Verification

- [ ] 4.1 Run focused doctor network tests
- [ ] 4.2 Run local teammate review for spec/test coverage
- [ ] 4.3 Run `go build ./...`
- [ ] 4.4 Run `go test ./...`
- [ ] 4.5 Run `git diff --check`
- [ ] 4.6 Sync main specs and run `openspec validate --all --strict`
- [ ] 4.7 Archive the completed OpenSpec change
