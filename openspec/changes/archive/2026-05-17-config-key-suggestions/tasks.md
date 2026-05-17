## 1. Tests
- [x] 1.1 Add failing `config get` tests for nearby key suggestions and generic key discovery hints.
- [x] 1.2 Add failing `config set` test for nearby key suggestions and no save on invalid path.

## 2. Implementation
- [x] 2.1 Add deterministic config key suggestion helpers in `internal/cli/configcmd`.
- [x] 2.2 Use the actionable error helper from `resolveConfigPath` and `setConfigPath`.
- [x] 2.3 Preserve existing behavior for valid paths, cleanup, and save failures.

## 3. Specs and Verification
- [x] 3.1 Sync main `config-cli-commands` spec.
- [x] 3.2 Validate the OpenSpec change in strict mode.
- [x] 3.3 Run focused config command tests.
- [x] 3.4 Run `go build ./...` and `go test ./...`.
- [x] 3.5 Run subagent-driven review.
- [x] 3.6 Archive the OpenSpec change and commit this scoped unit separately.
