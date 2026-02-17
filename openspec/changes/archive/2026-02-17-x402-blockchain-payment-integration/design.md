## Context

Lango is an AI agent platform with an existing security infrastructure: CryptoProvider interface, SecretsStore for encrypted data, RefStore for opaque references, RPCProvider for companion delegation, and CompositeCryptoProvider for primary/fallback. The payment system builds on top of these existing patterns to add blockchain USDC payments without exposing private keys to the agent.

The target chain is Base (L2 on Ethereum) with Base Sepolia as the default testnet. USDC is the payment token (ERC-20 with 6 decimal places).

## Goals / Non-Goals

**Goals:**
- Agent can send USDC payments, check balance, and view history through tools
- Private key NEVER exposed to agent context (stored encrypted, zeroed after signing)
- Spending limits enforced per-transaction and daily
- X402 HTTP 402 payment challenges can be parsed and auto-paid
- Feature flag gated (`payment.enabled=false` by default)
- Maximum reuse of existing security patterns (RPCProvider, CompositeCryptoProvider, SecretsStore)

**Non-Goals:**
- Token swaps or multi-token support (USDC only)
- On-chain transaction confirmation polling (submit-and-record only)
- Native ETH payments (ERC-20 USDC only)
- Multi-chain simultaneous operation (single chain per config)
- Fiat on-ramp integration

## Decisions

### 1. WalletProvider Interface (not CryptoProvider extension)
**Decision**: Create a separate `WalletProvider` interface rather than extending `CryptoProvider`.
**Rationale**: Wallet operations (Address, Balance, SignTransaction, SignMessage) are semantically different from crypto operations (Encrypt, Decrypt, Sign). Separate interfaces follow ISP and avoid bloating CryptoProvider.
**Alternatives**: Extending CryptoProvider — rejected because it would force all crypto implementations to implement wallet methods.

### 2. SecretsStore for Private Key Storage
**Decision**: Store the wallet private key as an encrypted secret via the existing SecretsStore under key `wallet.privatekey`.
**Rationale**: Reuses the proven encryption-at-rest infrastructure. The key is only loaded into memory for signing operations and immediately zeroed.
**Alternatives**: Separate keystore file — rejected as it duplicates SecretsStore functionality.

### 3. ERC-20 Transfer via Raw ABI Encoding
**Decision**: Manually ABI-encode `transfer(address,uint256)` rather than using abigen or a full ABI library.
**Rationale**: Only one function signature is needed. Manual encoding is simpler, avoids code generation, and has zero additional dependencies.
**Alternatives**: abigen — rejected as overkill for a single function call.

### 4. EIP-1559 Transactions
**Decision**: Use EIP-1559 (type 2) transactions with dynamic fee estimation.
**Rationale**: Base L2 supports EIP-1559 natively. Dynamic fees prevent overpayment and stuck transactions.

### 5. Ent-Based Transaction Log
**Decision**: Use an Ent schema (PaymentTx) for transaction logging rather than a separate database.
**Rationale**: Reuses the existing Ent/SQLite infrastructure. Provides typed queries, indexing, and migration support.

### 6. Spending Limits from Transaction Records
**Decision**: Calculate daily spending by summing PaymentTx records rather than maintaining a separate counter.
**Rationale**: Single source of truth. No counter drift. Accurate even after crashes or restarts.

## Risks / Trade-offs

- **[Transaction finality]** → Transactions are recorded as "submitted" not "confirmed". Finality checking is a future enhancement. Mitigation: conservative spending limits prevent overspend.
- **[RPC endpoint reliability]** → Single RPC endpoint per config. Mitigation: config supports changing the RPC URL; future work could add multiple endpoints.
- **[Gas estimation accuracy]** → Gas estimation may be inaccurate for congested networks. Mitigation: 2x baseFee multiplier provides headroom.
- **[Key compromise]** → If SecretsStore passphrase is compromised, wallet key is exposed. Mitigation: same risk profile as existing crypto keys; RPC wallet delegates to companion for higher security.
