# Smart Account Commands

Commands for managing ERC-7579 smart accounts with session keys, modules, and policies. Requires both smart account and payment to be enabled (`smartAccount.enabled = true`, `payment.enabled = true`).

```
lango account <subcommand>
```

!!! warning "Experimental Feature"
    The smart account system is experimental. Always verify transaction details, module addresses, and policy limits before executing on-chain operations.

---

## lango account info

Show smart account configuration and status including address, deployment state, installed modules, and paymaster status.

```
lango account info [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango account info
Smart Account Info
==================
Address:      0x1234abcd5678ef901234abcdef567890abcdef12
Deployed:     true
Owner:        0x5678abcd1234ef567890abcdef1234567890abcd
Chain ID:     84532
Entry Point:  0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789
Paymaster:    true

Installed Modules
-----------------
NAME                     TYPE         ADDRESS
LangoSessionValidator    validator    0xaaaa...
LangoSpendingHook        hook         0xbbbb...
LangoEscrowExecutor      executor     0xcccc...
```

When no modules are installed:

```bash
$ lango account info
Smart Account Info
==================
Address:      0x1234abcd...
...
No modules installed.
```

---

## lango account deploy

Deploy a new Safe smart account with the ERC-7579 adapter. If the account already exists, returns the existing account information.

```
lango account deploy [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango account deploy
Smart Account Deployed
  Address:     0x1234abcd5678ef901234abcdef567890abcdef12
  Deployed:    true
  Owner:       0x5678abcd1234ef567890abcdef1234567890abcd
  Chain ID:    84532
  Entry Point: 0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789
  Modules:     3

$ lango account deploy --output json
{
  "address": "0x1234abcd5678ef901234abcdef567890abcdef12",
  "isDeployed": true,
  "ownerAddress": "0x5678abcd1234ef567890abcdef1234567890abcd",
  "chainId": 84532,
  "entryPoint": "0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789",
  "moduleCount": 3
}
```

---

## lango account session list

List all session keys with their status, expiry, and spend limits.

```
lango account session list [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango account session list
ID          ADDRESS       PARENT    EXPIRES                    SPEND_LIMIT   STATUS
a1b2c3d4... 0x1234ab...   -         2026-03-09T14:30:00Z       1000000       active
e5f6a7b8... 0x5678cd...   a1b2c3d4  2026-03-08T10:00:00Z       unlimited     expired

$ lango account session list --output json
[
  {
    "id": "a1b2c3d4...",
    "address": "0x1234ab...",
    "expiresAt": "2026-03-09T14:30:00Z",
    "spendLimit": "1000000",
    "status": "active"
  }
]
```

When no sessions exist:

```bash
$ lango account session list
No session keys found.
```

---

## lango account session create

Create a new session key with delegated transaction signing permissions. Specify allowed targets, function selectors, spend limits, and duration.

```
lango account session create [--targets <addrs>] [--functions <selectors>] [--limit <wei>] [--duration <dur>] [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--targets` | string | `""` | Allowed target addresses (comma-separated) |
| `--functions` | string | `""` | Allowed function selectors (comma-separated) |
| `--limit` | string | `"0"` | Spend limit in wei |
| `--duration` | string | `"24h"` | Session duration (e.g., `1h`, `24h`, `168h`) |
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango account session create \
    --targets 0x036CbD53842c5426634e7929541eC2318f3dCF7e \
    --functions "0xa9059cbb" \
    --limit "5000000" \
    --duration 24h
Session Key Created
-------------------
ID:          a1b2c3d4-e5f6-7890-abcd-ef1234567890
Address:     0x9876fedc5432ba109876fedcba543210fedcba98
Targets:     0x036CbD53842c5426634e7929541eC2318f3dCF7e
Functions:   0xa9059cbb
Spend Limit: 5000000 wei
Expires:     2026-03-09T14:30:00Z
Created:     2026-03-08T14:30:00Z
```

---

## lango account session revoke

Revoke a specific session key by ID, or revoke all active session keys with `--all`.

```
lango account session revoke [session-id] [--all]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--all` | bool | `false` | Revoke all active session keys |

**Example:**

```bash
$ lango account session revoke a1b2c3d4-e5f6-7890-abcd-ef1234567890
Session key a1b2c3d4-e5f6-7890-abcd-ef1234567890 revoked.

$ lango account session revoke --all
All active session keys revoked.
```

!!! tip
    Either a session ID or the `--all` flag is required. The command will return an error if neither is provided.

---

## lango account module list

List all registered ERC-7579 modules including their name, type, address, and version.

```
lango account module list [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango account module list
NAME                     TYPE         ADDRESS                                      VERSION
LangoSessionValidator    validator    0xaaaa1234567890abcdef1234567890abcdef1234    1.0.0
LangoSpendingHook        hook         0xbbbb1234567890abcdef1234567890abcdef1234    1.0.0
LangoEscrowExecutor      executor     0xcccc1234567890abcdef1234567890abcdef1234    1.0.0

$ lango account module list --output json
[
  {
    "name": "LangoSessionValidator",
    "type": "validator",
    "address": "0xaaaa1234567890abcdef1234567890abcdef1234",
    "version": "1.0.0"
  }
]
```

When no modules are registered:

```bash
$ lango account module list
No modules registered.
```

---

## lango account module install

Install an ERC-7579 module on the smart account. Requires specifying the module type.

```
lango account module install <module-address> [--type validator|executor|fallback|hook]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--type` | string | `validator` | Module type (`validator`, `executor`, `fallback`, or `hook`) |

**Example:**

```bash
$ lango account module install 0xdddd1234567890abcdef1234567890abcdef1234 --type executor
Module installed successfully.
  Address:  0xdddd1234567890abcdef1234567890abcdef1234
  Type:     executor
  Tx Hash:  0xaabb1234...
```

!!! danger "On-Chain Operation"
    Module installation submits an on-chain transaction through the ERC-4337 bundler. Verify the module address and type before proceeding.

---

## lango account policy show

Show the current harness policy configuration for the smart account, including spending limits, allowed targets, and risk score requirements.

```
lango account policy show [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango account policy show
Harness Policy
==============
Account:            0x1234abcd5678ef901234abcdef567890abcdef12
Max Tx Amount:      5000000
Daily Limit:        50000000
Monthly Limit:      500000000
Auto-Approve Below: 100000
Required Risk Score: 0.80
Allowed Targets:    3 addresses
Allowed Functions:  2 selectors

$ lango account policy show --output json
{
  "account": "0x1234abcd5678ef901234abcdef567890abcdef12",
  "hasPolicy": true,
  "maxTxAmount": "5000000",
  "dailyLimit": "50000000",
  "monthlyLimit": "500000000",
  "autoApproveBelow": "100000",
  "allowedTargets": ["0xaaaa...", "0xbbbb...", "0xcccc..."],
  "allowedFunctions": ["0xa9059cbb", "0x095ea7b3"],
  "requiredRiskScore": 0.80
}
```

When no policy is set:

```bash
$ lango account policy show
Harness Policy
==============
Account: 0x1234abcd...
Status:  No policy set

Use 'lango account policy set' to configure limits.
```

---

## lango account policy set

Set harness policy spending limits. At least one limit flag must be provided. Updates the existing policy or creates a new one.

```
lango account policy set [--max-tx <wei>] [--daily <wei>] [--monthly <wei>]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--max-tx` | string | `""` | Maximum per-transaction amount in wei |
| `--daily` | string | `""` | Daily spending limit in wei |
| `--monthly` | string | `""` | Monthly spending limit in wei |

**Example:**

```bash
$ lango account policy set \
    --max-tx "5000000" \
    --daily "50000000" \
    --monthly "500000000"
Policy Updated
--------------
Account:       0x1234abcd5678ef901234abcdef567890abcdef12
Max Tx Amount: 5000000
Daily Limit:   50000000
Monthly Limit: 500000000
```

!!! tip
    All limit values are specified in wei. For USDC (6 decimals), `1000000` wei equals 1.00 USDC.

---

## lango account paymaster status

Show paymaster configuration and approval status, including provider type and RPC endpoint.

```
lango account paymaster status [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango account paymaster status
Paymaster Status
  Enabled:       true
  Provider:      pimlico
  Provider Type: pimlico
  RPC URL:       https://api.pimlico.io/v2/...
  Token:         0x036CbD53842c5426634e7929541eC2318f3dCF7e
  Paymaster:     0x00000000009726632680AF5d2E20f3c706e2F00e
  Policy ID:     sp_my_policy_id

$ lango account paymaster status --output json
{
  "enabled": true,
  "provider": "pimlico",
  "rpcURL": "https://api.pimlico.io/v2/...",
  "tokenAddress": "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
  "paymasterAddress": "0x00000000009726632680AF5d2E20f3c706e2F00e",
  "policyId": "sp_my_policy_id",
  "providerType": "pimlico"
}
```

---

## lango account paymaster approve

Approve the paymaster to spend USDC from the smart account. This is required before the paymaster can sponsor gas in USDC. Submits an ERC-20 `approve` transaction through the ERC-4337 bundler.

```
lango account paymaster approve [--amount <usdc>] [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--amount` | string | `"1000.00"` | USDC amount to approve (or `"max"` for unlimited) |
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango account paymaster approve --amount 1000.00
Paymaster USDC Approval Submitted
  Token:     0x036CbD53842c5426634e7929541eC2318f3dCF7e
  Paymaster: 0x00000000009726632680AF5d2E20f3c706e2F00e
  Amount:    1000.00 USDC
  Tx Hash:   0xaabb1234...

$ lango account paymaster approve --amount max --output json
{
  "token": "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
  "paymaster": "0x00000000009726632680AF5d2E20f3c706e2F00e",
  "amount": "max",
  "txHash": "0xccdd5678..."
}
```

!!! danger "On-Chain Operation"
    Paymaster approval submits a USDC `approve` transaction on-chain. Using `--amount max` grants unlimited spending approval. Verify the paymaster address before proceeding.
