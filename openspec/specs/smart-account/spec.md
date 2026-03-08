# Smart Account Specification

## ADDED Requirements

### R1: Solidity ERC-7579 Modules

Three on-chain modules implementing ERC-7579 interfaces:

1. **LangoSessionValidator** (TYPE_VALIDATOR): Validates UserOperation signatures against registered session keys and their policies (targets, functions, spend limits, expiry). Session key registration/revocation by account owner.

2. **LangoSpendingHook** (TYPE_HOOK): Pre/post execution hook enforcing per-session and global spending limits (per-tx, daily, cumulative). Tracks spend per session key with daily reset.

3. **LangoEscrowExecutor** (TYPE_EXECUTOR): Batched escrow operations (approve + createDeal + deposit) in a single UserOp via IERC7579Account.execute().

### R2: Core Go Types & Interfaces

Foundation types in `internal/smartaccount/`:
- `AccountManager` interface (GetOrDeploy, Info, InstallModule, UninstallModule, Execute)
- `SessionKey`, `SessionPolicy`, `ModuleInfo`, `UserOperation`, `ContractCall` structs
- 13 sentinel errors + `PolicyViolationError` custom type

### R3: Session Key Management

Package `internal/smartaccount/session/`:
- `Store` interface with in-memory implementation
- `Manager`: Create (ECDSA keypair), encrypt via CryptoProvider callback, register on-chain, sign UserOps, revoke (cascade children)
- Hierarchical: Master → Task sessions with policy intersection
- Lifecycle: Start/Stop, expired key cleanup

### R4: Policy Engine

Package `internal/smartaccount/policy/`:
- `HarnessPolicy`: MaxTxAmount, DailyLimit, MonthlyLimit, AllowedTargets, AllowedFunctions
- `Validator.Check()`: Pre-flight validation against policy + spend tracker
- `Engine`: Per-account policy management, risk-driven generation via callback
- `MergePolicies()`: Intersection of master + task policies

### R5: Account Manager & Bundler Client

- `Factory`: Compute counterfactual Safe address (CREATE2), deploy via Safe factory
- `Manager`: GetOrDeploy, InstallModule, UninstallModule, Execute via bundler
- `bundler.Client`: JSON-RPC for eth_sendUserOperation, eth_estimateUserOperationGas, eth_getUserOperationReceipt

### R6: Module Registry

Package `internal/smartaccount/module/`:
- `Registry`: Register/List/Get module descriptors
- `ABIEncoder`: Encode installModule/uninstallModule calldata (ERC-7579)
- Pre-registered: LangoSessionValidator, LangoSpendingHook, LangoEscrowExecutor

### R7: ABI Bindings

Package `internal/smartaccount/bindings/`:
- Typed clients for SessionValidator, SpendingHook, EscrowExecutor, Safe7579
- Uses `contract.ContractCaller` pattern (same as escrow hub)

### R8: Configuration

`SmartAccountConfig` in config types:
- Enabled, FactoryAddress, EntryPointAddress, Safe7579Address, BundlerURL
- Session: MaxDuration, DefaultGasLimit, MaxActiveKeys
- Modules: SessionValidatorAddress, SpendingHookAddress, EscrowExecutorAddress

### R9: Wallet Extension

`UserOpSigner` interface in wallet package:
- `SignUserOp(ctx, userOpHash, entryPoint, chainID) ([]byte, error)`
- `LocalUserOpSigner` implementation using ECDSA with Ethereum personal_sign

### R10: App Wiring & Agent Tools

- `wiring_smartaccount.go`: `initSmartAccount()` with callback-based cross-package wiring
- 10 agent tools: smart_account_deploy, smart_account_info, session_key_create/list/revoke, session_execute, policy_check, module_install/uninstall, spending_status
- Registered under "smartaccount" catalog category

### R11: CLI Commands

`lango account` command group:
- `deploy`, `info`, `session create/list/revoke`, `module list/install`, `policy show/set`
- All support `--output json|table` format

### R13: Session Key Paymaster Allowlist

The `SessionPolicy` struct SHALL include an `allowedPaymasters` field (address array). When non-empty, `validateUserOp` SHALL enforce that the paymaster address in `paymasterAndData` is in the allowlist. Empty array = all paymasters allowed (backward compatible). Short paymasterAndData (< 20 bytes) skips the check. `_setSession` persists the array.

### R12: Economy Integration

Callback-based integrations (no direct smartaccount imports):
- `budget.OnChainTracker`: Tracks per-session spending from on-chain data
- `risk.PolicyAdapter`: Converts risk assessments to session policy recommendations
- `sentinel.SessionGuard`: Revokes/restricts sessions on sentinel alerts
