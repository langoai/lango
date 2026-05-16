# Spec: Smart Account ABI Correctness

## Purpose

Capability spec for smart-account-abi. See requirements below for scope and behavior contracts.

## Requirements

### REQ-1: SessionValidator ABI must include allowedPaymasters

The `SessionValidatorABI` Go constant SHALL include `allowedPaymasters` (address[]) as the 8th tuple field in both `registerSessionKey` and `getSessionKeyPolicy` methods, matching `LangoSessionValidator.sol`.

#### Scenario: Session policy encodes allowedPaymasters
- **WHEN** a `SessionPolicy` with `allowedPaymasters` is registered on-chain
- **THEN** the ABI tuple SHALL encode all 8 fields correctly

### REQ-2: SpendingHook ABI must match LangoSpendingHook.sol

The Go binding SHALL expose:
- `setLimits(uint256, uint256, uint256)` — not the old `setLimit(address, uint256)`
- `getConfig(address) → (uint256, uint256, uint256)` — not `getLimit`
- `getSpendState(address, address) → (uint256, uint256, uint256)` — not `getSpentAmount`
- `resetSpentAmount` MUST be removed (does not exist on-chain)

#### Scenario: SetLimits uses the updated ABI signature
- **WHEN** `SetLimits` is called with per-tx=100, daily=1000, cumulative=10000
- **THEN** the correct ABI-encoded transaction SHALL be submitted

#### Scenario: GetConfig returns the three-limit struct
- **WHEN** `GetConfig` is called for an account address
- **THEN** it SHALL return `SpendingConfig{PerTxLimit, DailyLimit, CumulativeLimit}`

### REQ-3: UserOperation hash must follow ERC-4337 v0.7

The `computeUserOpHash()` function SHALL pack gas fields into `accountGasLimits` and `gasFees` 32-byte words per the PackedUserOperation spec.

#### Scenario: accountGasLimits packs verification and call gas
- **WHEN** `verificationGasLimit=100000` and `callGasLimit=200000` are hashed
- **THEN** `accountGasLimits` SHALL pack them into one 32-byte word with verification gas in the upper 128 bits

### REQ-4: Safe initializer must use proper ABI encoding

`buildSafeInitializer()` SHALL ABI-encode a `Safe.setup()` call with owners, threshold, fallback handler, and 7579 adapter address. The placeholder concatenation SHALL be replaced.

#### Scenario: Safe initializer uses ABI-encoded setup call
- **WHEN** `buildSafeInitializer()` constructs the initializer payload
- **THEN** it SHALL ABI-encode the `Safe.setup()` call instead of concatenating placeholders

### REQ-5: Nonce must be fetched from chain

`submitUserOp()` SHALL call `GetNonce()` to fetch the current account nonce, not use hardcoded `big.NewInt(0)`.

#### Scenario: submitUserOp fetches nonce from chain
- **WHEN** `submitUserOp()` prepares a user operation
- **THEN** it SHALL call `GetNonce()` instead of using a hardcoded zero nonce

### REQ-6: No duplicate ABI constants

`Safe7579ABI` SHALL be defined in exactly one location (`bindings/safe7579.go`). The duplicate in `factory.go` SHALL be removed.

#### Scenario: Safe7579ABI has a single canonical definition
- **WHEN** the smart-account ABI bindings are compiled
- **THEN** `Safe7579ABI` SHALL exist only in `bindings/safe7579.go`
- **AND** no duplicate constant SHALL remain in `factory.go`
