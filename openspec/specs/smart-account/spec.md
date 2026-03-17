# Smart Account Specification

## Purpose

ERC-7579 modular smart account support — Safe proxy deployment, session key management, module registry, policy engine, and paymaster integration.

## Requirements

### Requirement: Solidity ERC-7579 modules
The system SHALL provide three on-chain modules implementing ERC-7579 interfaces:
1. **LangoSessionValidator** (TYPE_VALIDATOR): Validates UserOperation signatures against registered session keys and their policies (targets, functions, spend limits, expiry). Session key registration/revocation by account owner.
2. **LangoSpendingHook** (TYPE_HOOK): Pre/post execution hook enforcing per-session and global spending limits (per-tx, daily, cumulative). Tracks spend per session key with daily reset.
3. **LangoEscrowExecutor** (TYPE_EXECUTOR): Batched escrow operations (approve + createDeal + deposit) in a single UserOp via IERC7579Account.execute().

#### Scenario: Module type compliance
- **WHEN** the on-chain modules are deployed
- **THEN** each module SHALL implement the corresponding ERC-7579 module type interface

### Requirement: Core Go types and interfaces
The system SHALL provide foundation types in `internal/smartaccount/`:
- `AccountManager` interface (GetOrDeploy, Info, InstallModule, UninstallModule, Execute)
- `SessionKey`, `SessionPolicy`, `ModuleInfo`, `UserOperation`, `ContractCall` structs
- 13 sentinel errors + `PolicyViolationError` custom type

#### Scenario: AccountManager interface completeness
- **WHEN** the AccountManager interface is defined
- **THEN** it SHALL include GetOrDeploy, Info, InstallModule, UninstallModule, and Execute methods

### Requirement: Session key management
The system SHALL provide session key management in `internal/smartaccount/session/`:
- `Store` interface with in-memory implementation
- `Manager`: Create (ECDSA keypair), encrypt via CryptoProvider callback, register on-chain, sign UserOps, revoke (cascade children)
- Hierarchical: Master → Task sessions with policy intersection
- Lifecycle: Start/Stop, expired key cleanup

#### Scenario: Session key creation and signing
- **WHEN** a session key is created via Manager.Create()
- **THEN** it SHALL generate an ECDSA keypair and store it encrypted via the CryptoProvider callback

#### Scenario: Hierarchical session revocation
- **WHEN** a master session key is revoked
- **THEN** all child task sessions SHALL be cascade-revoked

### Requirement: Policy engine
The system SHALL provide a policy engine in `internal/smartaccount/policy/`:
- `HarnessPolicy`: MaxTxAmount, DailyLimit, MonthlyLimit, AllowedTargets, AllowedFunctions
- `Validator.Check()`: Pre-flight validation against policy + spend tracker
- `Engine`: Per-account policy management, risk-driven generation via callback
- `MergePolicies()`: Intersection of master + task policies

#### Scenario: Policy validation
- **WHEN** a transaction is checked against a policy
- **THEN** the validator SHALL enforce all policy constraints (amount limits, allowed targets, allowed functions)

### Requirement: Account manager and bundler client
The Factory SHALL compute counterfactual Safe address (CREATE2) and deploy via Safe factory. `IsDeployed()` SHALL check for contract code at the address using `ethclient.CodeAt()` and return true if the code length is greater than zero. The Manager SHALL use `wallet.SignTransaction()` (raw sign) for UserOp hash signing. The bundler client SHALL provide JSON-RPC for eth_sendUserOperation, eth_estimateUserOperationGas, eth_getUserOperationReceipt. `GetNonce()` SHALL retrieve the nonce from the EntryPoint contract via eth_call to `EntryPoint.getNonce(address, uint192 key=0)` using selector `0x35567e1a`. `GetGasFees()` SHALL calculate MaxFeePerGas as `2 * baseFee + priorityFee`. The bundler client SHALL use v0.7 field format: `initCode` split into `factory`+`factoryData`, `paymasterAndData` split into `paymaster`+`paymasterVerificationGasLimit`+`paymasterPostOpGasLimit`+`paymasterData`.

#### Scenario: GetNonce uses EntryPoint
- **WHEN** GetNonce is called
- **THEN** it SHALL call EntryPoint.getNonce(address, 0) via eth_call, not eth_getTransactionCount

#### Scenario: v0.7 field format
- **WHEN** a UserOp is serialized for the bundler
- **THEN** initCode SHALL be split into factory + factoryData, and paymasterAndData SHALL be split into paymaster + gas limits + data

### Requirement: Module registry
The system SHALL provide a module registry in `internal/smartaccount/module/`:
- `Registry`: Register/List/Get module descriptors
- `ABIEncoder`: Encode installModule/uninstallModule calldata (ERC-7579)
- Pre-registered: LangoSessionValidator, LangoSpendingHook, LangoEscrowExecutor

#### Scenario: Pre-registered modules
- **WHEN** the module registry is initialized with module addresses
- **THEN** LangoSessionValidator, LangoSpendingHook, and LangoEscrowExecutor SHALL be pre-registered

### Requirement: ABI bindings
The system SHALL provide typed clients in `internal/smartaccount/bindings/` for SessionValidator, SpendingHook, EscrowExecutor, and Safe7579 using the `contract.ContractCaller` pattern.

#### Scenario: Typed client contract calls
- **WHEN** a binding client method is called
- **THEN** it SHALL use ContractCaller.Read() or ContractCaller.Write() with the correct ABI encoding

### Requirement: Configuration
`SmartAccountConfig` SHALL include: Enabled, FactoryAddress, EntryPointAddress, SafeSingletonAddress, Safe7579Address, FallbackHandler, BundlerURL, Session (MaxDuration, DefaultGasLimit, MaxActiveKeys), Modules (SessionValidatorAddress, SpendingHookAddress, EscrowExecutorAddress).

#### Scenario: Config fields present
- **WHEN** SmartAccountConfig is defined
- **THEN** it SHALL include all required fields with mapstructure and json tags

### Requirement: Wallet extension
The system SHALL provide a `UserOpSigner` interface in the wallet package with `SignUserOp(ctx, userOpHash, entryPoint, chainID) ([]byte, error)` and a `LocalUserOpSigner` implementation using ECDSA.

#### Scenario: UserOp signing
- **WHEN** SignUserOp is called
- **THEN** it SHALL sign the userOpHash using ECDSA with Ethereum personal_sign format

### Requirement: App wiring and agent tools
The system SHALL provide `initSmartAccount()` in `wiring_smartaccount.go` with callback-based cross-package wiring and 12 agent tools registered under "smartaccount" catalog category.

#### Scenario: Tool registration
- **WHEN** smart account is enabled and initialized
- **THEN** 12 tools SHALL be registered: smart_account_deploy, smart_account_info, session_key_create/list/revoke, session_execute, policy_check, module_install/uninstall, spending_status, paymaster_status, paymaster_approve

### Requirement: CLI commands
The system SHALL provide `lango account` command group with deploy, info, session create/list/revoke, module list/install, policy show/set. All SHALL support `--output json|table` format.

#### Scenario: CLI output formats
- **WHEN** a CLI command is run with --output json
- **THEN** it SHALL output valid JSON

### Requirement: Economy integration
The system SHALL provide callback-based integrations (no direct smartaccount imports):
- `budget.OnChainTracker`: Tracks per-session spending from on-chain data
- `risk.PolicyAdapter`: Converts risk assessments to session policy recommendations
- `sentinel.SessionGuard`: Revokes/restricts sessions on sentinel alerts

#### Scenario: Sentinel guard revocation
- **WHEN** a critical sentinel alert is published
- **THEN** the session guard SHALL invoke the revoke callback

### Requirement: Session key paymaster allowlist
The `SessionPolicy` struct SHALL include an `AllowedPaymasters` field (address array). When non-empty, `validateUserOp` SHALL enforce that the paymaster address in `paymasterAndData` is in the allowlist. Empty array means all paymasters are allowed.

#### Scenario: Paymaster allowlist enforcement
- **WHEN** a UserOp is validated with a non-empty paymaster allowlist
- **THEN** the validator SHALL reject UserOps whose paymaster is not in the list

### Requirement: CREATE2 address computation
The Factory SHALL compute the counterfactual address using the correct CREATE2 formula: `keccak256(0xff ++ factory ++ deploymentSalt ++ keccak256(proxyCreationCode ++ abi.encode(singleton)))`. The proxyCreationCode SHALL be fetched from the factory contract's `proxyCreationCode()` view function and cached.

#### Scenario: Correct initCodeHash computation
- **WHEN** ComputeAddress is called
- **THEN** the initCodeHash SHALL be `keccak256(proxyCreationCode ++ singletonPadded)` where singletonPadded is the singleton address left-padded to 32 bytes

#### Scenario: proxyCreationCode caching
- **WHEN** ComputeAddress is called multiple times
- **THEN** the proxyCreationCode SHALL be fetched from the factory contract only once and cached for subsequent calls

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

### Requirement: Session key UserOp hash correctness
The session key manager SHALL use the same UserOp hash algorithm as the EntryPoint contract (ComputeUserOpHash) instead of raw byte concatenation. The session manager SHALL accept entryPoint address and chainID via options.

#### Scenario: Session key signature matches EntryPoint
- **WHEN** a UserOp is signed with a session key
- **THEN** the signing digest SHALL match ComputeUserOpHash(op, entryPoint, chainID)

### Requirement: lango account exec guard
The `blockLangoExec` function SHALL include a guard entry for `lango account` that redirects the agent to the built-in smart account tools.

#### Scenario: Agent attempts lango account CLI
- **WHEN** the agent attempts to run `lango account deploy` or any `lango account` subcommand via exec
- **THEN** `blockLangoExec` SHALL return a message listing all smart account tool names

### Requirement: Init logging with config hints
The `initSmartAccount()` function SHALL include actionable configuration hints in its log messages when initialization is skipped.

#### Scenario: Smart account disabled
- **WHEN** `cfg.SmartAccount.Enabled` is false
- **THEN** the log message SHALL include a "fix" field with the command to enable it

#### Scenario: Payment components missing
- **WHEN** payment components are nil
- **THEN** the log message SHALL include a "fix" field listing required payment config keys

### Requirement: Safe Proxy Deployment
The Factory SHALL pass the Safe L2 singleton address (not the Safe7579 adapter) as the `_singleton` parameter to `createProxyWithNonce()`. The Safe7579 adapter SHALL be passed as the `to` parameter in `Safe.setup()` for delegate-call initialization.

#### Scenario: Deploy with correct singleton
- **WHEN** Factory.Deploy() is called
- **THEN** the createProxyWithNonce() call SHALL use singletonAddr as the singleton parameter
- **AND** the Safe.setup() initializer SHALL use safe7579Addr as the delegate-call target

#### Scenario: ComputeAddress uses correct singleton
- **WHEN** Factory.ComputeAddress() is called
- **THEN** the CREATE2 formula SHALL use singletonAddr in the proxy initCode hash

### Requirement: Safe Singleton Configuration
The system SHALL support a `SafeSingletonAddress` config field for specifying the Safe L2 singleton implementation address. When empty, it SHALL default to `0x29fcB43b46531BcA003ddC8FCB67FFE91900C762` (Safe L2 v1.4.1).

#### Scenario: Default singleton address
- **WHEN** SmartAccountConfig.SafeSingletonAddress is empty
- **THEN** the system SHALL use Safe L2 v1.4.1 as the default

#### Scenario: TUI settings field
- **WHEN** the settings TUI smart account form is displayed
- **THEN** a "Safe Singleton" field SHALL be visible with the default address as placeholder

### Requirement: Bundler Revert Reason Extraction
The bundler client SHALL extract and decode revert reasons from JSON-RPC error responses, supporting Error(string) and Panic(uint256) selectors.

#### Scenario: Error(string) revert
- **WHEN** a bundler JSON-RPC error contains an Error(string) ABI-encoded payload in the data field
- **THEN** the error message SHALL include the decoded revert string

#### Scenario: Panic(uint256) revert
- **WHEN** a bundler JSON-RPC error contains a Panic(uint256) payload
- **THEN** the error message SHALL include the panic description

#### Scenario: Missing or unparseable data
- **WHEN** a bundler JSON-RPC error has no data field or undecodable data
- **THEN** the error message SHALL omit the revert reason without crashing

### Requirement: Contract Caller Revert Diagnostics
The contract caller SHALL extract revert reasons from go-ethereum errors and replay failed transactions as eth_call to obtain revert data.

#### Scenario: go-ethereum DataError extraction
- **WHEN** a go-ethereum RPC error implements the dataError interface
- **THEN** the error message SHALL include the decoded revert reason from ErrorData()

#### Scenario: Gas estimation revert fallback
- **WHEN** EstimateGas fails without a direct DataError
- **THEN** the system SHALL replay as eth_call to extract the revert reason

#### Scenario: Receipt-level revert with eth_call replay
- **WHEN** a confirmed transaction has receipt.Status == 0
- **THEN** the system SHALL replay the call as eth_call at the revert block number and include the revert reason

### Requirement: Tool Execution Failure Logging
The system SHALL log tool execution failures at WARN level for server-side visibility.

#### Scenario: ADK tool handler error
- **WHEN** a tool call returns an error
- **THEN** a WARN-level log SHALL be emitted with tool name, agent name, and error details

#### Scenario: Event bus tool failure logging
- **WHEN** a ToolExecutedEvent with Success=false is published on the event bus
- **THEN** the observability subscriber SHALL log the failure at WARN level
