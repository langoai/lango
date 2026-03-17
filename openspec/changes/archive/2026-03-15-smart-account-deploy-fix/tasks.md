## 1. Singleton/Safe7579 Separation (Root Cause Fix)

- [x] 1.1 Add `SafeSingletonAddress` field to `SmartAccountConfig` in `internal/config/types_smartaccount.go`
- [x] 1.2 Separate `singletonAddr` from `safe7579Addr` in `Factory` struct in `internal/smartaccount/factory.go`
- [x] 1.3 Update `NewFactory` signature to accept both `singletonAddr` and `safe7579Addr` as separate parameters
- [x] 1.4 Update `Deploy()` to pass `singletonAddr` to `createProxyWithNonce()` and `safe7579Addr` to `buildSafeInitializer()`
- [x] 1.5 Update `ComputeAddress()` to use `singletonAddr` in CREATE2 formula

## 2. Config and Wiring

- [x] 2.1 Resolve singleton address with default Safe L2 v1.4.1 in `initSmartAccount()` in `internal/app/wiring_smartaccount.go`
- [x] 2.2 Same singleton resolution in CLI deps at `internal/cli/smartaccount/deps.go`
- [x] 2.3 Add TUI field for Safe Singleton in `internal/cli/settings/forms_smartaccount.go`
- [x] 2.4 Add `sa_singleton_address` state update case in `internal/cli/tuicore/state_update.go`

## 3. Revert Reason Decoding

- [x] 3.1 Add `Data *json.RawMessage` field to `jsonrpcError` in `internal/smartaccount/bundler/types.go`
- [x] 3.2 Add `RevertReason()` method to `jsonrpcError` in `internal/smartaccount/bundler/types.go`
- [x] 3.3 Create `internal/smartaccount/bundler/revert.go` with `DecodeRevertReason()`, `decodeErrorString()`, `decodePanicCode()`
- [x] 3.4 Update `call()` in `internal/smartaccount/bundler/client.go` to include revert reason in error messages

## 4. Contract Caller Revert Extraction

- [x] 4.1 Add `dataError` interface and `extractRevertReason()` in `internal/contract/caller.go`
- [x] 4.2 Add `replayForRevertReason()` method for eth_call replay at revert block
- [x] 4.3 Use `extractRevertReason()` in `Read()` for go-ethereum errors
- [x] 4.4 Use `extractRevertReason()` + `replayForRevertReason()` in `Write()` for gas estimation failures
- [x] 4.5 Use `replayForRevertReason()` in `Write()` for receipt-level reverts (status=0)

## 5. Tool Failure Logging

- [x] 5.1 Add WARN log in `adaptToolWithOptions` handler in `internal/adk/tools.go` when tool call returns error
- [x] 5.2 Subscribe to `ToolExecutedEvent` in `internal/app/wiring_observability.go` to log failed tool events

## 6. Tests

- [x] 6.1 Create `internal/smartaccount/bundler/revert_test.go` with table-driven tests for `DecodeRevertReason()`
- [x] 6.2 Update `NewFactory` calls in `internal/smartaccount/factory_test.go`
- [x] 6.3 Update `NewFactory` calls in `internal/smartaccount/manager_test.go`

## 7. Verification

- [x] 7.1 Run `go build ./...` and verify zero errors
- [x] 7.2 Run `go test ./internal/smartaccount/... ./internal/contract/... ./internal/adk/...` and verify all pass
