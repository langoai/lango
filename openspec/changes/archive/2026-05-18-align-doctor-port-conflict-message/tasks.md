## 1. Planning

- [x] 1.1 Create focused OpenSpec artifacts for doctor port conflict wording
- [x] 1.2 Validate and commit planning artifacts

## 2. Tests First

- [x] 2.1 Add failing doctor server port test for occupied-port message

## 3. Implementation

- [x] 3.1 Change occupied-port failure message to `Port <port> in use`

## 4. Review And Verification

- [x] 4.1 Run focused doctor network tests
- [x] 4.2 Run local teammate review for spec/test coverage
- [x] 4.3 Run `go build ./...`
- [x] 4.4 Run `go test ./...`
- [x] 4.5 Run `git diff --check`
- [x] 4.6 Sync main specs and run `openspec validate --all --strict`
- [x] 4.7 Archive the completed OpenSpec change
