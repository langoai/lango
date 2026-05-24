## 1. Planning

- [x] 1.1 Create focused OpenSpec artifacts for doctor network address formatting hardening
- [x] 1.2 Validate and commit planning artifacts

## 2. Tests First

- [x] 2.1 Add failing doctor server port test for IPv6 host formatting
- [x] 2.2 Add doctor server port coverage for bracketed IPv6 host normalization

## 3. Implementation

- [x] 3.1 Route `NetworkCheck` listen address formatting through `gatewayaddr.ListenAddress`
- [x] 3.2 Remove now-unused direct formatting dependency

## 4. Review And Verification

- [x] 4.1 Run focused doctor network tests
- [x] 4.2 Run subagent or local teammate review for spec/test coverage
- [x] 4.3 Run `go build ./...`
- [x] 4.4 Run `go test ./...`
- [x] 4.5 Run `git diff --check`
- [x] 4.6 Sync main specs and run `openspec validate --all --strict`
- [x] 4.7 Archive the completed OpenSpec change
