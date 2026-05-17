# Tasks

## 1. Planning

- [x] 1.1 Audit Telegram Bot API HTTP client construction and existing timeout behavior.
- [x] 1.2 Add focused OpenSpec artifacts for bounding the default Telegram HTTP client timeout.
- [x] 1.3 Validate the OpenSpec change before implementation.
- [x] 1.4 Commit the OpenSpec planning artifacts separately.

## 2. Regression Test

- [ ] 2.1 Add a test that requires the default Telegram HTTP client timeout to be finite and greater than long-poll timeout.
- [ ] 2.2 Add a test that preserves injected `Config.HTTPClient` unchanged.
- [ ] 2.3 Confirm the focused tests fail before implementation.

## 3. Implementation

- [ ] 3.1 Add an internal default Telegram HTTP client timeout.
- [ ] 3.2 Route `telegram.New` through the client resolver.
- [ ] 3.3 Avoid adding public configuration or changing long-poll/download timeout behavior.

## 4. Review

- [ ] 4.1 Request teammate review for reliability and test quality.
- [ ] 4.2 Address actionable findings before archiving.

## 5. Verification

- [ ] 5.1 Run focused Telegram channel tests.
- [ ] 5.2 Run `go build ./...`.
- [ ] 5.3 Run `go test ./...`.
- [ ] 5.4 Run `openspec validate --all --strict`.
- [ ] 5.5 Run `git diff --check`.
- [ ] 5.6 Archive the change after verification.
