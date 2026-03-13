## 1. toolparam Package

- [x] 1.1 Create `internal/toolparam/extract.go` with RequireString, OptionalString, OptionalInt, OptionalBool, StringSlice, ErrMissingParam
- [x] 1.2 Create `internal/toolparam/response.go` with Response type, StatusResponse, ListResponse builders
- [x] 1.3 Create `internal/toolparam/extract_test.go` with table-driven tests for all functions

## 2. Tool Handler Migration

- [x] 2.1 Migrate `internal/app/tools_workspace.go` to use toolparam helpers
- [x] 2.2 Migrate `internal/app/tools_escrow.go` to use toolparam helpers
- [x] 2.3 Migrate `internal/app/tools_economy.go` to use toolparam helpers
- [x] 2.4 Migrate `internal/app/tools_p2p.go` to use toolparam helpers
- [x] 2.5 Migrate `internal/app/tools_automation.go` to use toolparam helpers
- [x] 2.6 Migrate `internal/app/tools_smartaccount.go` to use toolparam helpers
- [x] 2.7 Migrate `internal/app/tools_contract.go` to use toolparam helpers
- [x] 2.8 Migrate `internal/app/tools_browser.go` to use toolparam helpers
- [x] 2.9 Migrate `internal/app/tools_data.go`, `tools_meta.go`, `tools_team.go`, `tools_sentinel.go`

## 3. Domain Errors

- [x] 3.1 Create `internal/economy/escrow/errors.go` with ErrNotFunded, ErrInvalidStatus
- [x] 3.2 Create `internal/p2p/workspace/errors.go` with ErrWorkspaceNotFound
- [x] 3.3 Create `internal/cli/workflow/errors.go` with ErrWorkflowDisabled
- [x] 3.4 Update `ent_store.go`, `manager.go`, `workflow.go` to use domain error sentinels

## 4. Constants Extraction

- [x] 4.1 Create `internal/config/constants.go` with exported validation maps
- [x] 4.2 Update `internal/config/loader.go` to reference validation constants
- [x] 4.3 Create `internal/economy/escrow/hub/methods.go` with method name constants
- [x] 4.4 Update hub `client.go` and `client_v2.go` to use method constants
- [x] 4.5 Add TransactionType constants to `internal/economy/escrow/types.go`

## 5. Sentinel & Hub Client Deduplication

- [x] 5.1 Convert AlertMetadata from map to typed struct in `sentinel/types.go`
- [x] 5.2 Update 5 detectors to use typed AlertMetadata in `sentinel/detector.go`
- [x] 5.3 Extract windowCounter struct and refactor RapidCreation/RepeatedDispute detectors
- [x] 5.4 Add writeMethod/readMethod helpers to HubClient and refactor all methods
- [x] 5.5 Update hub `client_test.go` for changed error message format

## 6. Verification

- [x] 6.1 Run `go build ./...` — zero errors
- [x] 6.2 Run `go test ./...` — all 103 packages pass
