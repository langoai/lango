# Smart Account Basics — ERC-4337 Smart Account + Session Keys

End-to-end integration test for Lango's ERC-4337 Smart Account deployment, session key management, and policy engine validation.

Spins up **1 Lango agent** and a local Ethereum node (Anvil) using Docker Compose, then verifies:

- Smart Account deployment (Safe-compatible)
- Session key creation and listing (master -> task keys)
- Policy engine validation
- Spending hook / spending status tracking

## Architecture

```
┌───────────────────┐
│   Agent           │
│   :18789          │
│   Smart Account   │
│   Session Keys    │
│   Policy Engine   │
└────────┬──────────┘
         │
    ┌────▼────┐
    │  Anvil  │  (chainId: 31337)
    │  :8545  │
    └─────────┘
```

## Configuration Highlights

| Setting | Value | Description |
|---------|-------|-------------|
| `smartAccount.enabled` | `true` | Enable ERC-4337 Smart Account features |
| `smartAccount.session.maxDuration` | `"24h"` | Maximum session key validity |
| `smartAccount.session.defaultGasLimit` | `500000` | Default gas limit for session operations |
| `smartAccount.session.maxActiveKeys` | `10` | Maximum concurrent session keys |
| `payment.limits.maxPerTx` | `"100.00"` | Maximum USDC per transaction |
| `payment.limits.autoApproveBelow` | `"50.00"` | Auto-approve payments under 50 USDC |
| `security.interceptor.headlessAutoApprove` | `true` | Auto-approve tool invocations in headless Docker mode |

> **Production Note**: Session key durations and spending limits are set high for testing convenience. In production, use shorter durations and lower limits appropriate to your use case.

## Prerequisites

- Docker & Docker Compose v2
- `cast` (from [Foundry](https://getfoundry.sh/)) — required for on-chain verification in the test script
- `curl` — for HTTP health/API checks

## Quick Start

```bash
# Build the Lango Docker image and start all services
make build up

# Run integration tests
make test

# Stop everything
make down
```

Or run everything in one command:

```bash
make all
```

## Services

| Service | Image | Purpose | Port |
|---------|-------|---------|------|
| `anvil` | `ghcr.io/foundry-rs/foundry` | Local EVM chain (chainId 31337) | 8545 |
| `setup` | `ghcr.io/foundry-rs/foundry` | Deploy MockUSDC + EntryPoint + Factory stubs | -- |
| `agent` | `lango:latest` | Smart Account agent | 18789 |

## Test Scenarios

1. **Health** — Agent responds to `GET /health`
2. **Contract Deployment** — MockUSDC, EntryPoint, and Factory stubs deployed on Anvil
3. **Smart Account Deploy** — Smart Account deployment tool available and functional
4. **Smart Account Info** — Smart Account info tool returns account details
5. **Session Key Operations** — Session key list tool returns key inventory
6. **Policy Check** — Policy engine validates target addresses
7. **Spending Status** — Spending hook tracks on-chain expenditure

## Anvil Test Accounts

| Role | Address | Private Key |
|------|---------|-------------|
| Agent | `0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266` | Account #0 |
| Deployer | `0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC` | Account #2 |

> **Note**: These are Anvil's well-known deterministic keys. Never use them on mainnet.

## REST API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/api/tools/execute` | POST | Execute a registered tool by name |

## Troubleshooting

```bash
# View all logs
make logs

# Check the agent
docker compose logs agent

# Manual health check
curl http://localhost:18789/health | jq .

# Check USDC balance on-chain
cast call $(cat /tmp/usdc-addr) "balanceOf(address)(uint256)" \
  0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266 --rpc-url http://localhost:8545

# Check contract deployment
docker compose exec agent cat /shared/entrypoint-address.txt
docker compose exec agent cat /shared/factory-address.txt
```
