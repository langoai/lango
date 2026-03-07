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
