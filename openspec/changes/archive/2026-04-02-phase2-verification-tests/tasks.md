## 1. Gateway Session Contract Tests (Unit 8)

- [x] 1.1 Add `TestChatMessage_UnauthenticatedGetsUniqueSessionKey` to `server_test.go`
- [x] 1.2 Add `TestChatMessage_AuthenticatedIgnoresClientSessionKey` to `server_test.go`
- [x] 1.3 Add `TestWebSocket_AbruptDisconnect` to `server_test.go`
- [x] 1.4 Add `TestHealth_AlwaysReturns200` to `server_test.go`
- [x] 1.5 Remove empty `gateway_test.go`

## 2. Payment Handler Tests (Unit 10)

- [x] 2.1 Create `internal/tools/payment/tools_test.go`
- [x] 2.2 Test base tool set (5 tools) and conditional tools (create_wallet, x402_fetch)
- [x] 2.3 Test safety level classification (dangerous vs safe)
- [x] 2.4 Test disabled interceptor does not add x402 tool
- [x] 2.5 Test handler parameter validation for payment_send

## 3. EventBus Contracts (Unit 14)

- [x] 3.1 Create `internal/app/eventbus_contracts_test.go`
- [x] 3.2 Test ToolExecutedEvent → collector.ToolExecutions increment
- [x] 3.3 Test PolicyDecisionEvent → collector.Policy.Blocks increment
- [x] 3.4 Test observe verdict → collector.Policy.Observes increment
- [x] 3.5 Test no events → counters unchanged
- [x] 3.6 Test multiple events accumulate correctly

## 4. Health Route Tests (Unit 15)

- [x] 4.1 Create `internal/app/routes_observability_test.go`
- [x] 4.2 Test /health/detailed with all healthy components
- [x] 4.3 Test /health/detailed with degraded component → worst-status aggregation
- [x] 4.4 Test nil registry → /health/detailed not registered (404)
- [x] 4.5 Test /metrics returns snapshot with expected fields

## 5. Verification

- [x] 5.1 Run `go build ./...` — passes
- [x] 5.2 Run `go test ./internal/gateway/... ./internal/app/... ./internal/tools/payment/...` — all pass
