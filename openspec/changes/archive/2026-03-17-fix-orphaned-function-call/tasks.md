## 1. Read-time Legacy Data Correction (Layer 2)

- [x] 1.1 Add role correction in `EventsAdapter.All()` for FunctionResponse messages stored with role "user"
- [x] 1.2 Add test `TestEventsAdapter_FunctionResponseUserRole` — verifies "user" role FunctionResponse is reconstructed correctly
- [x] 1.3 Add test `TestEventsAdapter_FunctionResponseToolRole` — verifies "tool" role FunctionResponse is unaffected

## 2. Write-time Role Correction (Layer 1)

- [x] 2.1 Add role correction in `AppendEvent` for FunctionResponse-only messages
- [x] 2.2 Add test `TestAppendEvent_FunctionResponseRoleCorrection` — verifies role corrected to "tool"
- [x] 2.3 Add test `TestAppendEvent_FunctionCallRoleUnchanged` — verifies FunctionCall role is not modified

## 3. Provider Boundary Defense (Layer 3)

- [x] 3.1 Implement `repairOrphanedFunctionCalls` function in `model.go`
- [x] 3.2 Integrate into `convertMessages` pipeline
- [x] 3.3 Add test `TestConvertMessages_OrphanedFunctionCall` — synthetic response injected
- [x] 3.4 Add test `TestConvertMessages_MatchedFunctionCall` — no injection when matched
- [x] 3.5 Add test `TestConvertMessages_PendingFunctionCallNotTouched` — pending calls untouched
- [x] 3.6 Add test `TestRepairOrphanedFunctionCalls_PartialResponse` — partial match handling

## 4. E2E Regression Test

- [x] 4.1 Add test `TestSessionRetry_OrphanedFunctionCallRegression` — full pipeline: AppendEvent → EventsAdapter.All() → convertMessages

## 5. Verification

- [x] 5.1 `go build ./...` passes
- [x] 5.2 `go test ./internal/adk/...` passes
