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

### Foundry Setup

```
contracts/
├── foundry.toml           # Solidity 0.8.24, optimizer 200 runs
├── src/
│   ├── LangoEscrowHub.sol
│   ├── LangoVault.sol
│   ├── LangoVaultFactory.sol
│   └── interfaces/
│       └── IERC20.sol
└── lib/
    └── forge-std/
```

Build and test:

```bash
cd contracts
forge build    # Compile contracts
forge test     # Run tests
```
