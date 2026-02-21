# Payment Commands

Commands for managing USDC blockchain payments on the Base L2 network. Payment must be enabled in configuration (`payment.enabled = true`). See the [Payments](../payments/index.md) section for detailed documentation.

```
lango payment <subcommand>
```

!!! warning "Experimental Feature"
    The payment system is experimental. Use with caution and always verify transaction details before sending.

---

## lango payment balance

Show the current USDC wallet balance, address, and network information.

```
lango payment balance [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango payment balance
Wallet Balance
  Balance:   25.50 USDC
  Address:   0x1234...abcd
  Network:   Base Sepolia (chain 84532)
```

---

## lango payment history

Show payment transaction history.

```
lango payment history [--json] [--limit N]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |
| `--limit` | int | `20` | Maximum number of transactions to show |

**Example:**

```bash
$ lango payment history
STATUS     AMOUNT     TO            METHOD   PURPOSE                  TX HASH       CREATED
confirmed  1.50 USDC  0x5678...     direct   API access fee           0xaabb...     2026-02-20 14:30
confirmed  0.50 USDC  0x9abc...     x402     Weather data query       0xccdd...     2026-02-20 13:15
pending    2.00 USDC  0xdef0...     direct   Document translation     0xeeff...     2026-02-20 12:00

$ lango payment history --limit 5 --json
```

---

## lango payment limits

Show configured spending limits and current daily usage.

```
lango payment limits [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango payment limits
Spending Limits
  Max Per Transaction:  1.00 USDC
  Max Daily:            10.00 USDC
  Spent Today:          3.50 USDC
  Remaining Today:      6.50 USDC
```

---

## lango payment info

Show wallet and payment system configuration details, including X402 protocol status.

```
lango payment info [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango payment info
Payment System Info
  Wallet Address:      0x1234...abcd
  Network:             Base Sepolia (chain 84532)
  Wallet Provider:     local
  USDC Contract:       0x036CbD53842c5426634e7929541eC2318f3dCF7e
  RPC URL:             https://sepolia.base.org
  X402 Auto-Intercept: disabled
  X402 Max Auto-Pay:   unlimited USDC
```

---

## lango payment send

Send a USDC payment to a recipient address. Requires `--to`, `--amount`, and `--purpose` flags. Prompts for confirmation unless `--force` is specified.

```
lango payment send --to <address> --amount <amount> --purpose <text> [--force] [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--to` | string | *required* | Recipient wallet address (`0x...`) |
| `--amount` | string | *required* | Amount in USDC (e.g., `"1.50"`) |
| `--purpose` | string | *required* | Human-readable purpose of the payment |
| `--force` | bool | `false` | Skip confirmation prompt |
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango payment send \
    --to 0x5678abcd1234ef567890abcdef1234567890abcd \
    --amount "1.50" \
    --purpose "API access fee"
Send 1.50 USDC to 0x5678abcd... on Base Sepolia?
Purpose: API access fee
Confirm [y/N]: y

Payment Submitted
  Status:    pending
  Tx Hash:   0xaabb1234...
  Amount:    1.50 USDC
  From:      0x1234abcd...
  To:        0x5678abcd...
  Network:   Base Sepolia (chain 84532)
```

!!! danger "Irreversible"
    Blockchain transactions cannot be reversed. Always verify the recipient address and amount before confirming.

!!! tip
    Use `--force` for non-interactive environments. Without it, the command requires confirmation and fails in non-interactive terminals.
