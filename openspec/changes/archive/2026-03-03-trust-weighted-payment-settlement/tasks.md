## 1. Config & Types

- [x] 1.1 Add TrustThresholds and SettlementConfig types to internal/config/types_p2p.go
- [x] 1.2 Create internal/p2p/paygate/trust.go with ReputationFunc, TrustConfig, DefaultTrustConfig
- [x] 1.3 Create internal/p2p/paygate/ledger.go with DeferredEntry and DeferredLedger

## 2. Payment Gate Core

- [x] 2.1 Add StatusPostPayApproved status and SettlementID field to Result in gate.go
- [x] 2.2 Add ReputationFn, TrustCfg to Config; add reputationFn, trustCfg, ledger to Gate struct
- [x] 2.3 Implement trust-based post-pay branch in Gate.Check()
- [x] 2.4 Remove Gate.SubmitOnChain() (replaced by settlement service)

## 3. Event Bus & Settlement

- [x] 3.1 Add ToolExecutionPaidEvent to internal/eventbus/events.go
- [x] 3.2 Add p2p_settlement enum value to ent schema payment_tx.go and run go generate
- [x] 3.3 Create internal/p2p/settlement/service.go with full lifecycle (subscribe, settle, retry, confirm)

## 4. Handler Integration

- [x] 4.1 Add payGateStatusPostPayApproved constant to protocol handler
- [x] 4.2 Add SettlementID field to PayGateResult
- [x] 4.3 Add eventBus field and SetEventBus() setter to Handler
- [x] 4.4 Update handleToolInvokePaid to handle postpay_approved and capture verifiedAuth
- [x] 4.5 Publish ToolExecutionPaidEvent after successful paid tool execution

## 5. Wiring

- [x] 5.1 Wire reputation function from repStore.GetScore to Gate via paygate.ReputationFunc
- [x] 5.2 Wire TrustThresholds config to TrustConfig
- [x] 5.3 Pass rpcClient through paymentComponents
- [x] 5.4 Create and wire settlement.Service to eventbus
- [x] 5.5 Wire handler.SetEventBus() with event bus instance
- [x] 5.6 Wire reputation recorder to settlement service
- [x] 5.7 Pass SettlementID through payGateAdapter

## 6. Tests

- [x] 6.1 Trust tier tests: high trust → postpay, medium → prepay, threshold → prepay, nil → prepay, error → prepay
- [x] 6.2 Ledger tests: Add/Settle/Pending/PendingByPeer/concurrent access
- [x] 6.3 Settlement service tests: defaults, subscribe, nil auth, wrong auth type, failure reputation
- [x] 6.4 Verify existing paygate and protocol handler tests still pass

## 7. Verification

- [x] 7.1 go build ./... passes
- [x] 7.2 go test ./internal/p2p/paygate/... passes
- [x] 7.3 go test ./internal/p2p/settlement/... passes
- [x] 7.4 go test ./internal/p2p/protocol/... passes
- [x] 7.5 go test ./... full regression passes
