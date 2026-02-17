## Why

Lango agents need the ability to make blockchain micropayments (USDC on Base L2) to interact with paid APIs and services autonomously. The X402 HTTP payment protocol is emerging as the standard for agent-to-service payments, and supporting it enables Lango to participate in the pay-per-use web. Critical requirement: the agent's private key must never be exposed to the LLM context.

## What Changes

- Add blockchain wallet management with encrypted private key storage (reusing SecretsStore)
- Add ERC-20 USDC transfer transaction building and submission
- Add spending limits enforcement (per-transaction and daily caps)
- Add 5 agent tools for payment operations (`payment_send`, `payment_balance`, `payment_history`, `payment_limits`, `payment_wallet_info`)
- Add X402 HTTP 402 payment protocol parsing and interception
- Add RPC wallet for companion app signing delegation (mirrors existing RPCProvider pattern)
- Add composite wallet with primary/fallback (mirrors CompositeCryptoProvider pattern)
- Add PaymentTx Ent schema for transaction audit trail
- Add payment configuration with feature flag (`payment.enabled`), network selection, and spending limits
- Wire payment tools into app startup, approval system, and multi-agent orchestration

## Capabilities

### New Capabilities
- `blockchain-wallet`: Wallet provider interface, local secp256k1 signing, RPC delegation, composite fallback, key isolation
- `payment-service`: USDC ERC-20 transfer building, transaction lifecycle (pending→submitted→confirmed→failed), spending limits
- `payment-tools`: Agent-facing tools for send, balance, history, limits, and wallet info
- `x402-protocol`: HTTP 402 payment challenge parsing, auto-intercept, payment proof headers

### Modified Capabilities
- `config-system`: New `PaymentConfig` block with network, limits, X402, and wallet provider settings
- `tool-safety-level`: `payment_send` registered as Dangerous, approval summary for payment actions
- `multi-agent-orchestration`: `payment_` prefix routed to Executor sub-agent

## Impact

- **New dependencies**: `github.com/ethereum/go-ethereum` (crypto, ethclient, types, common)
- **New packages**: `internal/wallet/`, `internal/payment/`, `internal/tools/payment/`, `internal/x402/`
- **Modified packages**: `internal/config/`, `internal/app/`, `internal/orchestration/`
- **New Ent schema**: `PaymentTx` with indexes on tx_hash, from_address, status, created_at, session_key
- **Feature flag**: All payment functionality gated behind `payment.enabled` (default: false)
- **Security**: Private keys stored encrypted via SecretsStore, zeroed after signing, never in agent context
