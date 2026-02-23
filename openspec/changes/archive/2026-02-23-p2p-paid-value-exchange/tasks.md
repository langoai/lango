## Tasks

### Step 1: Canonical USDC Contract Registry
- [x] Create `internal/payment/contracts/registry.go` with canonical addresses for chains 1, 8453, 84532, 11155111
- [x] Implement `LookupUSDC()`, `IsCanonical()`, `VerifyOnChain()` functions
- [x] Create `internal/payment/contracts/registry_test.go` with table-driven tests

### Step 2: EIP-3009 Authorization Builder
- [x] Create `internal/payment/eip3009/builder.go` with Authorization struct
- [x] Implement `NewUnsigned()`, `TypedDataHash()`, `Sign()`, `Verify()`, `EncodeCalldata()`
- [x] Create `internal/payment/eip3009/builder_test.go`

### Step 3: Owner Data Shield
- [x] Create `internal/p2p/firewall/owner_shield.go` with OwnerShield struct
- [x] Implement `ScanAndRedact()` with recursive map/slice scanning
- [x] Implement conversation key blocking (default true)
- [x] Create `internal/p2p/firewall/owner_shield_test.go`
- [x] Update `internal/p2p/firewall/firewall.go` to integrate OwnerShield

### Step 4: Reputation System
- [x] Create `internal/ent/schema/peer_reputation.go` Ent schema
- [x] Run `go generate ./internal/ent/...` for codegen
- [x] Create `internal/p2p/reputation/store.go` with RecordSuccess/Failure/Timeout
- [x] Implement `CalculateScore()` formula
- [x] Create `internal/p2p/reputation/store_test.go`
- [x] Add `ReputationChecker` callback to `internal/p2p/firewall/firewall.go`

### Step 5: Payment Gate
- [x] Create `internal/p2p/paygate/gate.go` with Check/SubmitOnChain/BuildQuote
- [x] Implement EIP-3009 auth parsing from JSON map
- [x] Implement ParseUSDC for decimal-to-smallest-unit conversion
- [x] Create `internal/p2p/paygate/gate_test.go`

### Step 6: Protocol Extension
- [x] Add `RequestPriceQuery`, `RequestToolInvokePaid` to messages.go
- [x] Add `StatusPaymentRequired` response status
- [x] Add `PriceQuoteResult` and `PaidInvokePayload` types
- [x] Add `handlePriceQuery()` and `handleToolInvokePaid()` to handler.go
- [x] Add `PayGateChecker` interface and `SetPayGate()` method
- [x] Add `QueryPrice()` and `InvokeToolPaid()` to remote_agent.go

### Step 7: ZK Proof Wiring
- [x] Create `initZKP()` function in wiring.go that compiles all 4 circuits
- [x] Wire ZK prover/verifier into handshake config
- [x] Wire ZK attestation function into firewall

### Step 8: Full Wiring
- [x] Add new imports to wiring.go (paygate, reputation, contracts, zkp, circuits, common, frontend, ent)
- [x] Extend `p2pComponents` struct with payGate and reputation fields
- [x] Change `initP2P()` signature to accept paymentComponents and ent.Client
- [x] Wire Owner Shield in initP2P
- [x] Wire Reputation system in initP2P
- [x] Wire Payment Gate in initP2P with PricingFunc
- [x] Wire PricingInfo on GossipCard
- [x] Create `payGateAdapter` to bridge paygate.Gate → protocol.PayGateChecker
- [x] Update `initP2P()` call in app.go with new parameters
- [x] Wire executor callback via closure after agent creation in app.go

### Step 9: Config Extensions
- [x] Add `P2PPricingConfig` struct to config/types.go
- [x] Add `OwnerProtectionConfig` struct to config/types.go
- [x] Add `MinTrustScore` field to P2PConfig

### Verification
- [x] `go build ./...` passes
- [x] `go test ./internal/p2p/...` passes
- [x] `go test ./internal/payment/...` passes
