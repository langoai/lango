---
title: USDC Payments
---

# USDC Payments

Lango provides a blockchain payment system for USDC on Base L2 (EVM). The agent can send USDC, check balances, view transaction history, and manage wallets through built-in payment tools.

!!! warning "Experimental"

    The payments system is under active development. Enable it with `payment.enabled: true`.

## Payment Tools

| Tool | Description | Safety Level |
|------|-------------|--------------|
| `payment_send` | Send USDC to recipient | Dangerous |
| `payment_balance` | Check wallet USDC balance | Safe |
| `payment_history` | View recent transactions | Safe |
| `payment_limits` | View spending limits and daily usage | Safe |
| `payment_wallet_info` | Show wallet address and network info | Safe |
| `payment_create_wallet` | Create new blockchain wallet | Dangerous |
| `x402_fetch` | HTTP request with automatic X402 payment | Dangerous |

!!! note "Tool Approval"

    Tools marked **Dangerous** require explicit user approval before execution unless auto-approve policies are configured. See [Tool Approval](../security/tool-approval.md) for details.

## Wallet Providers

| Provider | Description |
|----------|-------------|
| `local` | Default. Private key derived from encrypted secrets stored on disk. |
| `rpc` | Remote signer. Delegates signing to an external RPC endpoint. |
| `composite` | Tries `rpc` first, falls back to `local` if unavailable. |

## CLI Examples

Check wallet balance:

```bash
lango payment balance
```

View recent transactions:

```bash
lango payment history --limit 10
```

View spending limits and daily usage:

```bash
lango payment limits
```

Show wallet address and network info:

```bash
lango payment info
```

Send USDC to a recipient:

```bash
lango payment send --to 0x1234...abcd --amount 0.50 --purpose "API access"
```

Skip confirmation prompt with `--force`:

```bash
lango payment send --to 0x1234...abcd --amount 0.50 --purpose "API access" --force
```

JSON output for scripting:

```bash
lango payment balance --json
```

## Configuration

> **Settings:** `lango settings` → Payment

```json
{
  "payment": {
    "enabled": true,
    "walletProvider": "local",
    "network": {
      "chainId": 84532,
      "rpcUrl": "https://sepolia.base.org",
      "usdcContract": "0x..."
    },
    "limits": {
      "maxPerTx": 10.0,
      "maxDaily": 100.0,
      "autoApproveBelow": 0.10
    }
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `payment.enabled` | `bool` | `false` | Enable the payment system |
| `payment.walletProvider` | `string` | `local` | Wallet provider: `local`, `rpc`, or `composite` |
| `payment.network.chainId` | `int` | `84532` | Chain ID (`84532` = Base Sepolia, `8453` = Base mainnet) |
| `payment.network.rpcUrl` | `string` | -- | RPC endpoint URL for the target network |
| `payment.network.usdcContract` | `string` | -- | USDC token contract address |
| `payment.limits.maxPerTx` | `float64` | `10.0` | Maximum USDC per transaction |
| `payment.limits.maxDaily` | `float64` | `100.0` | Maximum USDC per 24-hour rolling window |
| `payment.limits.autoApproveBelow` | `float64` | `0.10` | Auto-approve threshold (no confirmation prompt) |

!!! tip "Testnet First"

    Start with Base Sepolia (`chainId: 84532`) for testing. Switch to Base mainnet (`chainId: 8453`) only after verifying your configuration. See the [Production Checklist](../deployment/production.md) for mainnet deployment guidance.

## Related

- [X402 Protocol](x402.md) -- Automatic HTTP 402 payment handling
- [Production Checklist](../deployment/production.md) -- Mainnet deployment guidance
- [Tool Approval](../security/tool-approval.md) -- Approval policies for dangerous tools
