# Proposal: ERC-7579 Modular Smart Account & Session Keys

## Problem

Lango currently operates on a **custody model** — the agent's wallet private key directly signs all blockchain transactions. This creates two critical problems:

1. **Security Risk**: If the agent is compromised, the attacker gains full control over the wallet
2. **Autonomy Bottleneck**: Every transaction requires either the master key or human approval

## Solution

Introduce **ERC-7579 Modular Smart Accounts** (Safe-based) with **Session Keys**, enabling "controlled autonomy" — agents operate within user-defined policy boundaries using time-limited, scope-restricted session keys. The master key is never exposed during routine operations.

## Approach

- **Account Type**: Safe (Gnosis) + ERC-7579 adapter (rhinestone/safe7579) — most mature, audited, Base-native
- **On-chain Scope**: Full on-chain modules (Validator + Executor + Hook)
- **Bundler**: External bundler RPC (Pimlico/Alchemy/StackUp) via `eth_sendUserOperation`
- **Dual Enforcement**: Policy checked BOTH off-chain (Go — fast rejection) AND on-chain (Solidity — tamper-proof)
- **Hierarchical Sessions**: Master Session (user-signed) → Task Session (agent-created within master bounds)
- **Callback Injection**: All cross-package wiring via typed function callbacks (no direct imports)

## Non-goals

- Custom account contracts (reuses Safe + modules only)
- Paymaster integration (future work)
- Multi-chain session key syncing
- On-chain governance for module upgrades

## Impact

- **Security**: Master key never exposed during routine agent operations
- **UX**: Agents can execute within policy bounds without per-tx approval
- **Extensibility**: Modular architecture allows adding new capabilities via ERC-7579 modules
