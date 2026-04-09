# Spec: Smart Account ABI Correctness

## Purpose

Capability spec for smart-account-abi. See requirements below for scope and behavior contracts.

## Requirements

### REQ-1: SessionValidator ABI must include allowedPaymasters

The `SessionValidatorABI` Go constant must include `allowedPaymasters` (address[]) as the 8th tuple field in both `registerSessionKey` and `getSessionKeyPolicy` methods, matching `LangoSessionValidator.sol`.

**Scenarios:**
- Given a SessionPolicy with allowedPaymasters set, when registered on-chain, then the tuple encodes all 8 fields correctly.

### REQ-2: SpendingHook ABI must match LangoSpendingHook.sol

The Go binding must expose:
- `setLimits(uint256, uint256, uint256)` — not the old `setLimit(address, uint256)`
- `getConfig(address) → (uint256, uint256, uint256)` — not `getLimit`
- `getSpendState(address, address) → (uint256, uint256, uint256)` — not `getSpentAmount`
- `resetSpentAmount` must be removed (does not exist on-chain)

**Scenarios:**
- Given per-tx=100, daily=1000, cumulative=10000, when `SetLimits` is called, then the correct ABI-encoded transaction is submitted.
- Given an account address, when `GetConfig` is called, then it returns `SpendingConfig{PerTxLimit, DailyLimit, CumulativeLimit}`.

### REQ-3: UserOperation hash must follow ERC-4337 v0.7

The `computeUserOpHash()` function must pack gas fields into `accountGasLimits` and `gasFees` 32-byte words per the PackedUserOperation spec.

**Scenarios:**
- Given verificationGasLimit=100000 and callGasLimit=200000, when hash is computed, then `accountGasLimits` packs them into a single 32-byte word with verification in upper 128 bits.

### REQ-4: Safe initializer must use proper ABI encoding

`buildSafeInitializer()` must ABI-encode a `Safe.setup()` call with owners, threshold, fallback handler, and 7579 adapter address. The placeholder concatenation must be replaced.

### REQ-5: Nonce must be fetched from chain

`submitUserOp()` must call `GetNonce()` to fetch the current account nonce, not use hardcoded `big.NewInt(0)`.

### REQ-6: No duplicate ABI constants

`Safe7579ABI` must be defined in exactly one location (`bindings/safe7579.go`). The duplicate in `factory.go` must be removed.
