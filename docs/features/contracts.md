---
title: Smart Contracts
---

# Smart Contracts

!!! warning "Experimental"

    Smart contract interaction is experimental. The tool interface and supported chains may change in future releases.

Lango supports direct EVM smart contract interaction with ABI caching. Agents can read on-chain state and send state-changing transactions through a unified tool interface.

## ABI Cache

Before calling a contract, its ABI must be loaded. Use `contract_abi_load` to pre-load and cache a contract ABI by address. Cached ABIs are reused across subsequent `contract_read` and `contract_call` invocations, avoiding repeated parsing.

## Read (View/Pure Calls)

The `contract_read` tool calls view or pure functions on a smart contract. These calls are free (no gas cost) and do not change on-chain state.

```
contract_read(address, abi, method, args?, chainId?)
```

Returns the decoded return value from the contract method.

## Write (State-Changing Calls)

The `contract_call` tool sends a state-changing transaction to a smart contract. These calls cost gas and may transfer ETH.

```
contract_call(address, abi, method, args?, value?, chainId?)
```

Returns the transaction hash and gas used.

## Agent Tools

| Tool | Safety | Description |
|------|--------|-------------|
| `contract_read` | Safe | Read data from a smart contract (view/pure call, no gas cost) |
| `contract_call` | Dangerous | Send a state-changing transaction to a smart contract (costs gas) |
| `contract_abi_load` | Safe | Pre-load and cache a contract ABI for faster subsequent calls |

## Configuration

Smart contract tools require payment to be enabled with a valid RPC endpoint:

```json
{
  "payment": {
    "enabled": true,
    "network": {
      "rpcURL": "https://mainnet.infura.io/v3/YOUR_KEY",
      "chainID": 1
    }
  }
}
```

See the [Contract CLI Reference](../cli/contract.md) for command documentation.

## Escrow Contracts

Lango includes Foundry-based Solidity contracts for on-chain escrow settlement between P2P agents.

### LangoEscrowHub

**Source:** `contracts/src/LangoEscrowHub.sol`

Master escrow hub for P2P agent deals. Holds multiple deals in a single contract, reducing deployment costs.

**Deal struct:** `buyer`, `seller`, `token`, `amount`, `deadline`, `status`, `workHash`

**States:** Created(0) → Deposited(1) → WorkSubmitted(2) → Released(3) / Refunded(4) / Disputed(5) → Resolved(6)

**Events:** `DealCreated`, `Deposited`, `WorkSubmitted`, `Released`, `Refunded`, `Disputed`, `DealResolved`

**Access control:**

| Modifier | Functions |
|----------|-----------|
| `onlyBuyer` | `deposit`, `release`, `refund` |
| `onlySeller` | `submitWork` |
| `onlyArbitrator` | `resolveDispute` |
| Either party | `dispute` |

### LangoVault

**Source:** `contracts/src/LangoVault.sol`

Individual vault per deal, designed as an EIP-1167 clone target. Same lifecycle as LangoEscrowHub but with `initialize()` instead of a constructor, enabling minimal proxy deployment.

**States:** Uninitialized(0) → Created(1) → Deposited(2) → WorkSubmitted(3) → Released(4) / Refunded(5) / Disputed(6) → Resolved(7)

**Events:** `VaultInitialized`, `Deposited`, `WorkSubmitted`, `Released`, `Refunded`, `Disputed`, `VaultResolved`

### LangoVaultFactory

**Source:** `contracts/src/LangoVaultFactory.sol`

EIP-1167 Minimal Proxy factory for LangoVault. Each call to `createVault()` clones the implementation contract and initializes the new vault with deal parameters.

**Events:** `VaultCreated`

### ERC-7579 Module Contracts

Lango deploys three custom ERC-7579 modules on the Safe smart account:

| Module | Type | Description |
|--------|------|-------------|
| **LangoSessionValidator** | Validator | Validates session key signatures and enforces per-session spending limits |
| **LangoSpendingHook** | Hook | Tracks on-chain spending per session key, enforces daily/monthly aggregate limits |
| **LangoEscrowExecutor** | Executor | Executes escrow operations (deposit, release, refund) through the smart account |

These modules are configured via `smartAccount.modules.*` keys and installed using `lango account module install`. See [Smart Accounts](smart-accounts.md) for details.

## Settlement Service

The settlement service (`internal/p2p/settlement/`) handles asynchronous on-chain settlement of P2P tool invocation payments. It subscribes to `ToolExecutionPaidEvent` from the event bus and submits `transferWithAuthorization` transactions (EIP-3009) to the USDC contract.

### Settlement Lifecycle

```
ToolExecutionPaidEvent ──► Create DB record (pending) ──► Build EIP-1559 tx ──► Sign via wallet ──► Submit with retry ──► Wait for confirmation
```

1. **Event subscription** -- Listens for `ToolExecutionPaidEvent` on the event bus
2. **Record creation** -- Creates a `PaymentTx` record in the database with status `pending`
3. **Transaction building** -- Encodes EIP-3009 `transferWithAuthorization` calldata, estimates gas, and constructs an EIP-1559 dynamic fee transaction
4. **Signing** -- Signs the transaction hash via the wallet provider
5. **Submission with retry** -- Sends the transaction with exponential backoff (default: 3 retries)
6. **Confirmation** -- Polls for the transaction receipt with exponential backoff (default timeout: 2 minutes)
7. **Reputation recording** -- Records success or failure against the peer's reputation score

Transaction nonces are serialized via a mutex to prevent nonce collisions across concurrent settlement goroutines.

### Configuration

| Setting | Default | Description |
|---------|---------|-------------|
| `settlement.receiptTimeout` | `2m` | Maximum time to wait for on-chain confirmation |
| `settlement.maxRetries` | `3` | Maximum submission retry attempts |

## Session Key Management

The session key system (`internal/smartaccount/session/`) manages ephemeral ECDSA keys scoped to specific policies for automated smart account operations.

### Key Hierarchy

Session keys support a parent-child hierarchy:

- **Master sessions** -- Root-level keys with full policy bounds (`parentID` is empty)
- **Task sessions** -- Child keys scoped within a parent's bounds, created with `intersectPolicies()` to enforce the tighter constraint for each field

### Session Policy

Each session key is constrained by a `SessionPolicy`:

| Field | Description |
|-------|-------------|
| `allowedTargets` | Contract addresses the key can interact with |
| `allowedFunctions` | 4-byte function selectors the key can call |
| `spendLimit` | Maximum cumulative spend allowed |
| `spentAmount` | Amount spent so far |
| `validAfter` / `validUntil` | Time window during which the key is valid |
| `allowedPaymasters` | Paymaster addresses the key can use |

### Key Lifecycle

| Operation | Description |
|-----------|-------------|
| `Create` | Generate ECDSA key pair, optionally encrypt private key material, register on-chain |
| `Get` / `List` | Retrieve session key metadata |
| `Revoke` | Mark key and all children as revoked, revoke on-chain if callback is set |
| `RevokeAll` | Revoke all active session keys |
| `SignUserOp` | Sign a UserOperation with a session key (decrypts private key material if encrypted) |
| `CleanupExpired` | Remove expired session keys from the store |

### Security

- Private key material can be encrypted at rest via `CryptoEncryptFunc` / `CryptoDecryptFunc` callbacks
- Maximum session duration is enforced (default: 24 hours)
- Maximum active keys limit is enforced (default: 10)
- Child sessions cannot exceed parent bounds

## Policy Engine

The policy engine (`internal/smartaccount/policy/`) provides off-chain pre-validation of contract calls before they are submitted on-chain.

### Harness Policy

The `HarnessPolicy` defines per-account constraints:

| Field | Description |
|-------|-------------|
| `maxTxAmount` | Maximum value per transaction |
| `dailyLimit` | Maximum daily aggregate spend |
| `monthlyLimit` | Maximum monthly aggregate spend |
| `allowedTargets` | Permitted contract addresses |
| `allowedFunctions` | Permitted function selectors |
| `requiredRiskScore` | Minimum risk score for approval |
| `autoApproveBelow` | Auto-approve transactions below this value |

### Spend Tracking

The `SpendTracker` maintains daily and monthly cumulative spend counters with automatic window resets:

- Daily counter resets after 24 hours
- Monthly counter resets after 30 days

### Policy Syncer

The `Syncer` synchronizes Go-side harness policies with on-chain `LangoSpendingHook` limits:

- `PushToChain` -- Writes Go-side policy to the SpendingHook contract
- `PullFromChain` -- Reads on-chain config and updates Go-side policy
- `DetectDrift` -- Compares Go-side and on-chain policies, reports differences

### Policy Merging

`MergePolicies(master, task)` produces the intersection of two policies by taking the tighter constraint for each field. This is used when creating task-scoped session keys within a master session.

## Module Registry

The module registry (`internal/smartaccount/module/`) manages available ERC-7579 module descriptors. Each module is described by:

| Field | Description |
|-------|-------------|
| `name` | Human-readable module name |
| `address` | On-chain contract address |
| `type` | Module type (validator, executor, fallback, hook) |
| `version` | Module version string |
| `initData` | Initialization data for module installation |

The registry supports listing by module type and is thread-safe for concurrent access.

## Paymaster Integration

The paymaster system (`internal/smartaccount/paymaster/`) enables gasless transactions via ERC-4337 paymaster sponsorship.

### Supported Providers

| Provider | Description |
|----------|-------------|
| **Alchemy** | Alchemy Gas Manager sponsorship |
| **Pimlico** | Pimlico verifying paymaster |
| **Circle** | Circle programmable wallets paymaster |

### Recovery

The `RecoverableProvider` wraps any paymaster provider with retry and fallback logic:

- **Transient errors** -- Retried with exponential backoff (default: 2 retries, 200ms base delay)
- **Permanent errors** -- Fail immediately without retry
- **Fallback mode** -- When retries are exhausted, either abort (`abort`) or fall back to direct gas payment (`direct`)

### Foundry Setup

```
contracts/
├── foundry.toml           # Solidity 0.8.24, optimizer 200 runs
├── src/
│   ├── LangoEscrowHub.sol
│   ├── LangoVault.sol
│   ├── LangoVaultFactory.sol
│   ├── interfaces/
│   │   └── IERC20.sol
│   └── modules/
│       ├── LangoSessionValidator.sol
│       ├── LangoSpendingHook.sol
│       ├── LangoEscrowExecutor.sol
│       └── ISessionValidator.sol
└── lib/
    └── forge-std/
```

Build and test:

```bash
cd contracts
forge build    # Compile contracts
forge test     # Run tests
```
