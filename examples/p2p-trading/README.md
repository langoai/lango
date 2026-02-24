# P2P Trading Integration Example

End-to-end integration test for Lango's P2P networking and USDC payment system.

Spins up **3 Lango agents** (Alice, Bob, Charlie) and a local Ethereum node (Anvil) using Docker Compose, then verifies:

- mDNS peer discovery
- P2P status and identity REST API
- DID derivation from wallet keys
- ERC-20 (MockUSDC) token transfer between agents

## Architecture

```
┌──────────┐     ┌──────────┐     ┌──────────┐
│  Alice    │◄───►│   Bob    │◄───►│ Charlie  │
│ :18789   │     │ :18790   │     │ :18791   │
│ P2P:9001 │     │ P2P:9002 │     │ P2P:9003 │
└────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │
     └────────┬───────┘────────────────┘
              │
         ┌────▼────┐
         │  Anvil  │  (chainId: 31337)
         │  :8545  │
         └─────────┘
```

## Configuration Highlights

The example agents are configured with the following approval and payment settings:

| Setting | Value | Description |
|---------|-------|-------------|
| `payment.limits.autoApproveBelow` | `"50.00"` | Auto-approve payments under 50 USDC without confirmation |
| `p2p.autoApproveKnownPeers` | `true` | Skip handshake approval for previously authenticated peers |
| `p2p.pricing.enabled` | `true` | Enable paid tool invocations between agents |
| `p2p.pricing.perQuery` | `"0.10"` | Default USDC price per tool query |
| `security.interceptor.headlessAutoApprove` | `true` | Auto-approve tool invocations in headless Docker mode |

> **Production Note**: The `autoApproveBelow` threshold is intentionally high (`50.00`) for testing convenience. In production, use a much lower value (e.g., `"0.10"`) and rely on interactive approval for larger amounts.

## Prerequisites

- Docker & Docker Compose v2
- `cast` (from [Foundry](https://getfoundry.sh/)) — required for balance checks in the test script
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

| Service   | Image                          | Purpose                          | Port  |
|-----------|--------------------------------|----------------------------------|-------|
| `anvil`   | `ghcr.io/foundry-rs/foundry`   | Local EVM chain (chainId 31337)  | 8545  |
| `setup`   | `ghcr.io/foundry-rs/foundry`   | Deploy MockUSDC + fund agents    | —     |
| `alice`   | `lango:latest`                 | Agent 1                          | 18789 |
| `bob`     | `lango:latest`                 | Agent 2                          | 18790 |
| `charlie` | `lango:latest`                 | Agent 3                          | 18791 |

## Test Scenarios

1. **Health** — All 3 agents respond to `GET /health`
2. **P2P Status** — `GET /api/p2p/status` returns peer ID and listen addresses
3. **P2P Discovery** — After 15s, each agent sees >= 2 peers via mDNS
4. **P2P Identity** — `GET /api/p2p/identity` returns a `did:lango:` DID
5. **USDC Balance** — On-chain `balanceOf` confirms 1000 USDC per agent
6. **Payment** — Alice sends 1.00 USDC to Bob; Bob's balance increases

## Anvil Test Accounts

| Agent   | Address                                      | Private Key |
|---------|----------------------------------------------|-------------|
| Alice   | `0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266` | Account #0  |
| Bob     | `0x70997970C51812dc3A010C7d01b50e0d17dc79C8` | Account #1  |
| Charlie | `0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC` | Account #2  |

> **Note**: These are Anvil's well-known deterministic keys. Never use them on mainnet.

## REST API Endpoints

| Endpoint             | Method | Description                     |
|----------------------|--------|---------------------------------|
| `/health`            | GET    | Health check                    |
| `/api/p2p/status`    | GET    | Peer ID, listen addrs, peer count |
| `/api/p2p/peers`     | GET    | List connected peers + addresses |
| `/api/p2p/identity`  | GET    | Local DID string                |
| `/api/p2p/reputation`| GET    | Peer trust score and history    |
| `/api/p2p/pricing`   | GET    | Tool pricing configuration      |

## Troubleshooting

```bash
# View all logs
make logs

# Check a specific agent
docker compose logs alice

# Manual API check
curl http://localhost:18789/api/p2p/status | jq .

# Check USDC balance on-chain
cast call $(cat /tmp/usdc-addr) "balanceOf(address)(uint256)" \
  0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266 --rpc-url http://localhost:8545
```
