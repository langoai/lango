# X402 V2 Design

## Architecture

```
Agent Tool (x402_fetch)
    │
    ▼
x402.Interceptor.HTTPClient()  ← lazy init, thread-safe
    │
    ▼
x402http.PaymentRoundTripper   ← SDK handles 402→sign→retry
    │
    ├─ BeforePaymentCreation hook → SpendingLimiter.Check()
    │
    ▼
evmclient.ExactEvmScheme       ← EIP-3009 / EIP-712 signing
    │
    ▼
evmsigners.ClientSigner         ← from LocalSignerProvider
```

## Key Decisions

### 1. SDK-First Approach
Use the Coinbase X402 Go SDK as-is rather than implementing protocol details ourselves. The SDK handles version detection (V1/V2), header encoding (Base64), and payment signing (EIP-712).

### 2. Spending Limits via Hooks
The SDK provides lifecycle hooks. We use `BeforePaymentCreationHook` to enforce:
- Per-transaction max auto-pay amount
- Daily spending limit via `SpendingLimiter`

### 3. Lazy Client Initialization
The X402-wrapped HTTP client is created on first use and cached. This avoids loading the private key until actually needed.

### 4. Separate from Direct Transfers
`payment_send` (direct ERC-20 transfer) and `x402_fetch` (automatic X402) are distinct tools:
- `payment_send`: User explicitly sends USDC to a recipient
- `x402_fetch`: Agent makes HTTP request; payment happens automatically on 402

### 5. Audit Trail
X402 payments are recorded in PaymentTx with `payment_method = "x402_v2"` for distinguishing from direct transfers in history and spending calculations.

## Unchanged Components
- Wallet providers (`internal/wallet/`) — reused as-is
- SpendingLimiter — reused by X402 interceptor hook
- TxBuilder — stays for direct `payment_send`
- Existing 6 payment tools — all remain
- CLI payment commands — unchanged
