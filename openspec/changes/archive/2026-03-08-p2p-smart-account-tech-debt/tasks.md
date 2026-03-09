# Tasks: P2P + Smart Account Technical Debt Resolution

## Phase A: ABI/Encoding Correctness

- [x] A1: Add `allowedPaymasters` to SessionValidator ABI and extend `SessionPolicy` struct with `SpentAmount`, `Active`, `AllowedPaymasters`
- [x] A2: Rewrite SpendingHook ABI with correct methods (`setLimits`, `getConfig`, `getSpendState`) and add `SpendingConfig`/`SpendState` types
- [x] A3: Rewrite `computeUserOpHash()` for ERC-4337 v0.7 PackedUserOperation format with `packGasValues`/`padTo32` helpers
- [x] A4: Implement proper `buildSafeInitializer()` with Safe.setup() ABI encoding and fix `ComputeAddress()` CREATE2 derivation
- [x] A5: Add `GetNonce()` to bundler client and wire into `submitUserOp()` replacing hardcoded zero nonce
- [x] A6: Remove duplicate `Safe7579ABI` from factory.go (use `bindings.Safe7579ABI`), create `scripts/check-abi.sh`

## Phase B: Security Fixes

- [x] B1: Add `escapePassphrase()` to dbmigrate and apply at 4 SQL interpolation sites
- [x] B2: Fix session key encryption — store hex-encoded ciphertext, pass key ID to decrypt
- [x] B3: Add `NonceCache` to `p2pComponents` for lifecycle management, wire default-deny `ApprovalFn` with reputation backfill
- [x] B4: Replace ZK challenge echo with actual ECDSA signature via `wp.SignMessage()`

## Phase C: Callback Wiring

- [x] C1: Wire `WithOnChainRegistration`/`WithOnChainRevocation` for session manager when SessionValidator configured, add `toOnChainPolicy()` converter
- [x] C2: Wire `OnChainTracker.SetCallback()` to budget engine `Record()` (replace log-only stub)
- [x] C3: Wire CardFn to handler, start gossip service, replace team invoke stub with real handler-based implementation
- [x] C4: Add 6 accessor methods to `smartAccountComponents`, expose via `App.SmartAccountComponents` field

## Phase D: Stub → Real Implementation

- [x] D1: Implement all CLI commands with real service calls via `smartAccountDeps` pattern (deploy, info, session, policy, module, paymaster)
- [x] D2: Create `PolicySyncer` with `PushToChain`, `PullFromChain`, `DetectDrift` methods
- [x] D3: Create `RecoverableProvider` with retry/fallback, add `IsTransient`/`IsPermanent`, wire into paymaster init

## Phase E: Integration Tests

- [x] E1: SmartAccount integration tests (6 tests: session lifecycle, paymaster 2-phase, policy enforcement, cumulative spend, encryption/decryption)
- [x] E2: P2P wiring tests (6 tests: nonce cache lifecycle, approval default-deny patterns)
- [x] E3: Cross-layer tests (10 tests: budget tracker sync, session guard revocation, policy syncer drift detection)
