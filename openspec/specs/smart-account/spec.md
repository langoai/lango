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

- `Factory`: Compute counterfactual Safe address (CREATE2), deploy via Safe factory. `IsDeployed()` SHALL check for contract code at the address using `ethclient.CodeAt()` and return true if the code length is greater than zero. Errors from `CodeAt()` SHALL be propagated to the caller.
- `Manager`: GetOrDeploy, InstallModule, UninstallModule, Execute via bundler. After `factory.Deploy()` succeeds, `GetOrDeploy()` SHALL call `IsDeployed()` to verify the contract actually exists on-chain. The manager SHALL use `wallet.SignTransaction()` (raw sign) for UserOp hash signing instead of `wallet.SignMessage()` (which internally applies keccak256), since `computeUserOpHash()` already returns a keccak256 digest.
- `bundler.Client`: JSON-RPC for eth_sendUserOperation, eth_estimateUserOperationGas, eth_getUserOperationReceipt. `GetNonce()` SHALL retrieve the nonce from the EntryPoint contract via eth_call to `EntryPoint.getNonce(address, uint192 key=0)` using selector `0x35567e1a` (SHALL NOT use `eth_getTransactionCount`). `GetGasFees()` SHALL retrieve EIP-1559 gas fee parameters from the network; `MaxFeePerGas` SHALL be calculated as `2 * baseFee + priorityFee`, with fallback to 1.5 gwei if `eth_maxPriorityFeePerGas` is not supported. The manager SHALL call `GetGasFees()` before gas estimation and set the fee parameters on the UserOp.

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

### Requirement: CREATE2 address computation
The Factory SHALL compute the counterfactual address using the correct CREATE2 formula: `keccak256(0xff ++ factory ++ deploymentSalt ++ keccak256(proxyCreationCode ++ abi.encode(singleton)))`. The proxyCreationCode SHALL be fetched from the factory contract's `proxyCreationCode()` view function and cached.

#### Scenario: Correct initCodeHash computation
- **WHEN** ComputeAddress is called
- **THEN** the initCodeHash SHALL be `keccak256(proxyCreationCode ++ singletonPadded)` where singletonPadded is the singleton address left-padded to 32 bytes

#### Scenario: proxyCreationCode caching
- **WHEN** ComputeAddress is called multiple times
- **THEN** the proxyCreationCode SHALL be fetched from the factory contract only once and cached for subsequent calls

#### Scenario: ComputeAddress requires context
- **WHEN** ComputeAddress is called
- **THEN** it SHALL accept a context.Context parameter and return (common.Address, error) to handle RPC failures

### Requirement: Deploy address resolution
The Factory.Deploy() SHALL first attempt to parse the actual deployed address from the contract call result. If the result does not contain a parseable address, it SHALL fall back to ComputeAddress.

#### Scenario: Deploy returns actual on-chain address
- **WHEN** the createProxyWithNonce call returns an address in the result data
- **THEN** Deploy SHALL return that address instead of computing it

#### Scenario: Deploy fallback to computed address
- **WHEN** the createProxyWithNonce result does not contain a parseable address
- **THEN** Deploy SHALL fall back to ComputeAddress with the corrected CREATE2 formula

### Requirement: UserOp receipt verification
The Manager.submitUserOp() SHALL wait for the UserOp to be included in an on-chain transaction by polling GetUserOperationReceipt(). It SHALL return the actual on-chain transaction hash instead of the UserOp hash.

#### Scenario: Successful UserOp confirmation
- **WHEN** a UserOp is submitted and included on-chain
- **THEN** submitUserOp SHALL return the on-chain transaction hash from the receipt

#### Scenario: UserOp reverted on-chain
- **WHEN** a UserOp is submitted but the receipt indicates failure (success=false)
- **THEN** submitUserOp SHALL return an error indicating on-chain revert

#### Scenario: UserOp receipt timeout
- **WHEN** no receipt is received within 2 minutes
- **THEN** submitUserOp SHALL return a timeout error

### Requirement: Session key UserOp hash correctness
The session key manager SHALL use the same UserOp hash algorithm as the EntryPoint contract (ComputeUserOpHash) instead of raw byte concatenation. The session manager SHALL accept entryPoint address and chainID via options.

#### Scenario: Session key signature matches EntryPoint
- **WHEN** a UserOp is signed with a session key
- **THEN** the signing digest SHALL match ComputeUserOpHash(op, entryPoint, chainID) which uses proper ABI encoding with 32-byte padding, keccak256 on variable-length fields, packed gas values, and double-hash with entryPoint+chainID

#### Scenario: Session manager configuration
- **WHEN** a session manager is created
- **THEN** it SHALL accept WithEntryPoint and WithChainID options for UserOp hash computation
