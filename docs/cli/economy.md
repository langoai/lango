# Economy Commands

Commands for managing P2P economy features including budget, risk, pricing, negotiation, and escrow. Economy must be enabled in configuration (`economy.enabled = true`).

```
lango economy <subcommand>
```

!!! warning "Experimental Feature"
    The P2P economy system is experimental. Use with caution and verify all economic parameters before enabling in production.

---

## lango economy budget status

Show budget configuration and allocation status.

```
lango economy budget status [--task-id <id>]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--task-id` | string | `""` | Task ID to check specific budget |

**Example:**

```bash
$ lango economy budget status
Budget Configuration:
  Default Max:      10.00 USDC
  Alert Thresholds: [0.5 0.8 0.95]
  Hard Limit:       enabled

$ lango economy budget status --task-id=task-1
Budget Configuration:
  Default Max:      10.00 USDC
  Alert Thresholds: [0.5 0.8 0.95]
  Hard Limit:       enabled

Task "task-1" budget: use 'lango serve' and economy_budget_status tool for live data
```

When economy is disabled:

```bash
$ lango economy budget status
Economy layer is disabled. Enable with economy.enabled=true
```

---

## lango economy risk status

Show risk assessment configuration including escrow thresholds and trust score tiers.

```
lango economy risk status
```

No additional flags.

**Example:**

```bash
$ lango economy risk status
Risk Configuration:
  Escrow Threshold: 5.00 USDC
  High Trust Score: 0.80
  Med Trust Score:  0.50
```

When economy is disabled:

```bash
$ lango economy risk status
Economy layer is disabled.
```

---

## lango economy pricing status

Show dynamic pricing configuration including discount rates and minimum price.

```
lango economy pricing status
```

No additional flags.

**Example:**

```bash
$ lango economy pricing status
Pricing Configuration:
  Trust Discount:  20%
  Volume Discount: 10%
  Min Price:       0.01 USDC
```

When pricing is disabled:

```bash
$ lango economy pricing status
Dynamic pricing is disabled.
```

---

## lango economy negotiate status

Show negotiation protocol configuration including round limits and auto-negotiation settings.

```
lango economy negotiate status
```

No additional flags.

**Example:**

```bash
$ lango economy negotiate status
Negotiation Configuration:
  Max Rounds:     5
  Timeout:        30s
  Auto Negotiate: true
  Max Discount:   30%
```

When negotiation is disabled:

```bash
$ lango economy negotiate status
Negotiation is disabled.
```

---

## lango economy escrow status

Show escrow service configuration including timeout, milestone limits, and dispute settings.

```
lango economy escrow status
```

No additional flags.

**Example:**

```bash
$ lango economy escrow status
Escrow Configuration:
  Default Timeout: 24h
  Max Milestones:  10
  Auto Release:    true
  Dispute Window:  48h
```

When escrow is disabled:

```bash
$ lango economy escrow status
Escrow is disabled.
```

---

## lango economy escrow list

Show escrow configuration summary including on-chain mode.

```
lango economy escrow list
```

No additional flags.

**Example:**

```bash
$ lango economy escrow list
Escrow Summary:
  On-Chain Escrow:  enabled
  Mode:             hub
  Hub Address:      0x1234...
  Auto Release:     false
  Default Timeout:  24h0m0s

Use 'lango economy escrow show' for detailed on-chain configuration.
```

When economy is disabled:

```bash
$ lango economy escrow list
Economy layer is disabled. Enable with economy.enabled=true
```

---

## lango economy escrow show

Show detailed on-chain escrow configuration including all contract addresses and settlement parameters.

```
lango economy escrow show [--id <escrow-id>]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--id` | string | `""` | Escrow ID to show (future use) |

**Example:**

```bash
$ lango economy escrow show
On-Chain Escrow Configuration:
  Enabled:              enabled
  Mode:                 hub
  Hub Address:          0x1234...
  Vault Factory:        (not set)
  Vault Implementation: (not set)
  Arbitrator:           0x5678...
  Token Address:        0x036CbD53842c5426634e7929541eC2318f3dCF7e
  Poll Interval:        15s

Settlement:
  Receipt Timeout:      2m0s
  Max Retries:          3
```

---

## lango economy escrow sentinel status

Show Security Sentinel engine status.

```
lango economy escrow sentinel status
```

No additional flags.

**Example:**

```bash
$ lango economy escrow sentinel status
Sentinel Engine:
  Status:  active (monitors on-chain escrow events)
  Mode:    hub

The sentinel engine runs within the application server.
Use 'lango serve' to start and 'lango economy escrow sentinel alerts'
(via agent tools) to view detected alerts.
```

When on-chain escrow is disabled:

```bash
$ lango economy escrow sentinel status
On-chain escrow is disabled. Sentinel monitors on-chain events.
```
