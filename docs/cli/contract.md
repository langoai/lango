# Contract Commands

Commands for interacting with EVM smart contracts. Requires the payment system to be enabled (`payment.enabled = true`).

```
lango contract <subcommand>
```

!!! warning "Experimental Feature"
    Contract interaction is experimental. Always verify contract addresses, method signatures, and ABI files before executing calls.

---

## lango contract read

Call a view/pure contract method (read-only, no gas required). Validates the ABI and method locally; live RPC queries require a running `lango serve` instance.

```
lango contract read --address <addr> --abi <file> --method <name> [--args <csv>] [--chain-id <id>] [--output]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--address` | string | *required* | Contract address (`0x...`) |
| `--abi` | string | *required* | Path to ABI JSON file |
| `--method` | string | *required* | Method name to call |
| `--args` | string | `""` | Comma-separated method arguments |
| `--chain-id` | int | from config | Chain ID override |
| `--output` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango contract read \
    --address 0x036CbD53842c5426634e7929541eC2318f3dCF7e \
    --abi ./erc20.json \
    --method balanceOf \
    --args 0x1234abcd5678ef901234abcdef567890abcdef12
Note: contract read requires a running RPC connection.
Use 'lango serve' and the contract_read agent tool for live queries.

Contract Read (validated)
  Address:  0x036CbD53842c5426634e7929541eC2318f3dCF7e
  Method:   balanceOf
  Args:     [0x1234abcd5678ef901234abcdef567890abcdef12]
  Chain ID: 84532
```

---

## lango contract call

Execute a state-changing transaction on a smart contract. Validates the ABI and method locally; live transactions require a running `lango serve` instance and wallet.

```
lango contract call --address <addr> --abi <file> --method <name> [--args <csv>] [--value <eth>] [--chain-id <id>] [--output]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--address` | string | *required* | Contract address (`0x...`) |
| `--abi` | string | *required* | Path to ABI JSON file |
| `--method` | string | *required* | Method name to call |
| `--args` | string | `""` | Comma-separated method arguments |
| `--value` | string | `""` | ETH value to send (e.g., `"0.01"`) |
| `--chain-id` | int | from config | Chain ID override |
| `--output` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango contract call \
    --address 0x036CbD53842c5426634e7929541eC2318f3dCF7e \
    --abi ./erc20.json \
    --method transfer \
    --args 0x5678abcd1234ef567890abcdef1234567890abcd,1000000
Note: contract call requires a running RPC connection and wallet.
Use 'lango serve' and the contract_call agent tool for live transactions.

Contract Call (validated)
  Address:  0x036CbD53842c5426634e7929541eC2318f3dCF7e
  Method:   transfer
  Args:     [0x5678abcd1234ef567890abcdef1234567890abcd 1000000]
  Chain ID: 84532
```

!!! danger "State-Changing"
    Contract calls may modify blockchain state and spend gas or tokens. Always verify the method, arguments, and value before executing.

---

## lango contract abi load

Parse and validate a contract ABI from a local JSON file. Caches the parsed ABI for subsequent read/call commands.

```
lango contract abi load --address <addr> --file <path> [--chain-id <id>] [--output]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--address` | string | *required* | Contract address (`0x...`) |
| `--file` | string | *required* | Path to ABI JSON file |
| `--chain-id` | int | from config | Chain ID override |
| `--output` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango contract abi load \
    --address 0x036CbD53842c5426634e7929541eC2318f3dCF7e \
    --file ./erc20.json
ABI Loaded
  Address:  0x036CbD53842c5426634e7929541eC2318f3dCF7e
  Chain ID: 84532
  Methods:  10
  Events:   3

$ lango contract abi load \
    --address 0x036CbD53842c5426634e7929541eC2318f3dCF7e \
    --file ./erc20.json \
    --output
{
  "address": "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
  "chainId": 84532,
  "events": 3,
  "methods": 10,
  "status": "loaded"
}
```
