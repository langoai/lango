# Paid Tool Marketplace

P2P paid tool invocations with prepaid/postpaid settlement on a local Ethereum chain.

Spins up **3 Lango agents** (Alice, Bob, Charlie) and a local Ethereum node (Anvil) using Docker Compose to demonstrate:

- Tool pricing configuration (per-query default + per-tool overrides)
- ERC-20 prepayment for tool invocations
- Trust-based post-pay settlement for high-reputation peers
- On-chain USDC settlement between agents

## Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│    Alice      │◄───►│     Bob      │◄───►│   Charlie    │
│  SELLER       │     │  BUYER       │     │  BUYER (hi)  │
│  :18789       │     │  :18790      │     │  :18791      │
│  P2P:9001     │     │  P2P:9002    │     │  P2P:9003    │
│               │     │              │     │              │
│  toolPrices:  │     │  Pays before │     │  Post-pay    │
│  search=0.25  │     │  invocation  │     │  (trust>0.8) │
│  browse=0.50  │     │              │     │              │
│  web   =0.15  │     │              │     │              │
│  review=1.00  │     │              │     │              │
└──────┬────────┘     └──────┬───────┘     └──────┬───────┘
       │                     │                    │
       └─────────┬───────────┘────────────────────┘
                 │
            ┌────▼────┐
            │  Anvil  │  (chainId: 31337)
            │  :8545  │
            └─────────┘
```

## Configuration Highlights

| Setting | Value | Description |
|---------|-------|-------------|
| `p2p.pricing.enabled` | `true` | Enable paid tool invocations |
| `p2p.pricing.perQuery` | `"0.10"` | Default USDC price per tool query |
| `p2p.pricing.toolPrices` | `{...}` | Per-tool price overrides (Alice only) |
| `p2p.pricing.trustThresholds.postPayMinScore` | `0.8` | Minimum trust score for deferred payment |
| `payment.limits.autoApproveBelow` | `"50.00"` | Auto-approve payments under 50 USDC |
| `security.interceptor.headlessAutoApprove` | `true` | Auto-approve tool invocations in headless mode |

> **Production Note**: The `autoApproveBelow` threshold is intentionally high for testing. In production, use a much lower value and rely on interactive approval.

## Prerequisites

- Docker & Docker Compose v2
- `cast` (from [Foundry](https://getfoundry.sh/)) -- required for on-chain balance checks
- `curl` -- for HTTP health/API checks

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

| Service   | Image                          | Purpose                              | Port  |
|-----------|--------------------------------|--------------------------------------|-------|
| `anvil`   | `ghcr.io/foundry-rs/foundry`   | Local EVM chain (chainId 31337)      | 8545  |
| `setup`   | `ghcr.io/foundry-rs/foundry`   | Deploy MockUSDC + fund agents        | --    |
| `alice`   | `lango:latest`                 | Seller (pricing enabled, tool prices)| 18789 |
| `bob`     | `lango:latest`                 | Buyer (standard prepay)              | 18790 |
| `charlie` | `lango:latest`                 | Buyer (high-trust, post-pay eligible)| 18791 |

## Test Scenarios

1. **Health Checks & Discovery** -- All 3 agents healthy and discover >= 2 peers via mDNS
2. **USDC Balances** -- On-chain `balanceOf` confirms 1000 USDC per agent after minting
3. **Pricing Configuration** -- Alice exposes tool-specific pricing via API
4. **P2P Identity & DID** -- Each agent has a `did:lango:` identity
5. **Reputation Baseline** -- Reputation endpoints available on all agents
6. **On-Chain Transfer** -- Bob sends 0.25 USDC to Alice (prepayment simulation)
7. **Post-Pay Settlement** -- Charlie settles 1.00 USDC deferred payment to Alice

## Anvil Test Accounts

| Agent   | Address                                      | Role           |
|---------|----------------------------------------------|----------------|
| Alice   | `0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266` | Seller         |
| Bob     | `0x70997970C51812dc3A010C7d01b50e0d17dc79C8` | Buyer          |
| Charlie | `0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC` | Buyer (hi-trust)|

> **Note**: These are Anvil's well-known deterministic keys. Never use them on mainnet.

## REST API Endpoints

| Endpoint             | Method | Description                       |
|----------------------|--------|-----------------------------------|
| `/health`            | GET    | Health check                      |
| `/api/p2p/status`    | GET    | Peer ID, listen addrs, peer count |
| `/api/p2p/peers`     | GET    | List connected peers + addresses  |
| `/api/p2p/identity`  | GET    | Local DID string                  |
| `/api/p2p/reputation`| GET    | Peer trust score and history      |
| `/api/p2p/pricing`   | GET    | Tool pricing configuration        |

## Troubleshooting

```bash
# View all logs
make logs

# Check a specific agent
docker compose logs alice

# Manual API check
curl http://localhost:18789/api/p2p/pricing | jq .

# Check USDC balance on-chain
cast call $(cat /tmp/usdc-addr) "balanceOf(address)(uint256)" \
  0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266 --rpc-url http://localhost:8545
```
