## 1. SmartAccount Config Validation (A1, A6)

- [x] 1.1 Add Validate() method to SmartAccountConfig in internal/config/types_smartaccount.go
- [x] 1.2 Insert Validate() call in internal/app/wiring_smartaccount.go before component creation
- [x] 1.3 Integrate Validate() in internal/cli/smartaccount/deps.go
- [x] 1.4 Write table-driven tests for Validate() in types_smartaccount_test.go

## 2. SessionGuard Lifecycle (A3)

- [x] 2.1 Add Stop() method and active-flag guard to SessionGuard.handleAlert()
- [x] 2.2 Add *lifecycle.Registry parameter to initSmartAccount signature
- [x] 2.3 Register SessionGuard with lifecycle registry at PriorityAutomation
- [x] 2.4 Update app.go call site to pass app.registry
- [x] 2.5 Add Stop() test in session_guard_test.go

## 3. Warning Logs (A2, A4, C1, F1)

- [x] 3.1 Add risk engine skip warning in wiring_smartaccount.go (A2)
- [x] 3.2 Add sentinel guard skip warning in wiring_smartaccount.go (A4)
- [x] 3.3 Add X402 secrets nil warning in wiring_payment.go (C1)
- [x] 3.4 Add observability sub-flag conflict warning in wiring_observability.go (F1)

## 4. Payment RPCURL Pre-validation (B1)

- [x] 4.1 Add empty RPCURL check before ethclient.Dial in wiring_payment.go

## 5. Disabled Category Registration (E1, A5, B2, D1)

- [x] 5.1 Add disabled category for browser
- [x] 5.2 Add disabled categories for crypto and secrets
- [x] 5.3 Add disabled category for meta (knowledge)
- [x] 5.4 Add disabled category for graph
- [x] 5.5 Add disabled category for rag
- [x] 5.6 Add disabled category for memory
- [x] 5.7 Add disabled category for agent_memory
- [x] 5.8 Add disabled categories for payment, contract, p2p, workspace
- [x] 5.9 Add disabled category for librarian
- [x] 5.10 Add disabled category for mcp
- [x] 5.11 Add disabled category for economy
- [x] 5.12 Add disabled category for observability
- [x] 5.13 Update smart account disabled description with required fields (A5)

## 6. Verification

- [x] 6.1 go build ./... passes
- [x] 6.2 go test ./... passes
