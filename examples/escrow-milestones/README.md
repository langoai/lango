# Escrow Milestones — On-Chain Escrow + Milestone Settlement

End-to-end integration test for Lango's on-chain escrow system with milestone-based fund release.

Spins up **2 Lango agents** (Alice=Buyer, Bob=Seller) and a local Ethereum node (Anvil) using Docker Compose, then verifies:

- EscrowHubV2 contract deployment and interaction
- MilestoneSettler and DirectSettler contract deployment
- Budget allocation and tracking
- Risk assessment configuration
- Milestone-based fund release
- P2P discovery between buyer and seller

## Architecture

```
┌──────────────┐          ┌──────────────┐
│    Alice     │◄────────►│     Bob      │
│   (Buyer)    │   P2P    │   (Seller)   │
│   :18789     │          │   :18790     │
│  P2P:9001    │          │  P2P:9002    │
└──────┬───────┘          └──────┬───────┘
       │                         │
       └────────┬────────────────┘
                │
   ┌────────────▼────────────┐
   │         Anvil           │
   │    (chainId: 31337)     │
   │         :8545           │
   │                         │
   │  MockUSDC               │
   │  EscrowHubV2Stub        │
   │  MilestoneSettlerStub   │
   │  DirectSettlerStub      │
   └─────────────────────────┘
```

## Configuration Highlights

| Setting | Value | Description |
|---------|-------|-------------|
| `economy.enabled` | `true` | Enable economy subsystem |
| `economy.budget.defaultMax` | `"50.00"` | Maximum budget per session (USDC) |
| `economy.budget.hardLimit` | `true` | Enforce hard budget cap |
| `economy.escrow.enabled` | `true` | Enable escrow functionality |
| `economy.escrow.maxMilestones` | `10` | Maximum milestones per escrow |
| `economy.escrow.autoRelease` | `true` | Auto-release on milestone completion |
| `economy.escrow.onChain.enabled` | `true` | Use on-chain escrow contracts |
| `economy.escrow.onChain.mode` | `"hub"` | Use EscrowHubV2 mode |
| `economy.risk.escrowThreshold` | `"5.00"` | Minimum amount requiring escrow |
| `economy.risk.highTrustScore` | `0.8` | Trust score threshold for reduced escrow |

> **Production Note**: The `autoRelease` flag is enabled for testing convenience. In production, require explicit approval for milestone releases with appropriate multi-sig or governance controls.

## Prerequisites

- Docker & Docker Compose v2
- `cast` (from [Foundry](https://getfoundry.sh/)) — required for on-chain balance checks in the test script
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
| `setup` | `ghcr.io/foundry-rs/foundry` | Deploy MockUSDC + escrow contracts + fund agents | — |
| `alice` | `lango:latest` | Buyer agent | 18789 |
| `bob` | `lango:latest` | Seller agent | 18790 |

## Test Scenarios

1. **Health** — Both agents respond to `GET /health`
2. **Contract Deployment** — MockUSDC, EscrowHubV2, MilestoneSettler, DirectSettler deployed
3. **P2P Discovery** — After 15s, each agent discovers the other via mDNS
4. **DID Identity** — `GET /api/p2p/identity` returns a `did:lango:` DID for each agent
5. **USDC Balance** — On-chain `balanceOf` confirms 1000 USDC per agent
6. **Escrow Contract Verification** — EscrowHubV2 and MilestoneSettler return correct versions
7. **On-Chain Escrow Simulation** — Alice funds escrow, milestone releases to Bob
8. **Budget Tracking** — Alice's balance reflects escrow funding deduction
9. **Economy Configuration** — Budget, milestone, and risk settings are correctly applied

## Anvil Test Accounts

| Agent | Address | Private Key |
|-------|---------|-------------|
| Alice (Buyer) | `0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266` | Account #0 |
| Bob (Seller) | `0x70997970C51812dc3A010C7d01b50e0d17dc79C8` | Account #1 |
| Deployer | `0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC` | Account #2 |

> **Note**: These are Anvil's well-known deterministic keys. Never use them on mainnet.

## Troubleshooting

```bash
# View all logs
make logs

# Check a specific agent
docker compose logs alice

# Manual API check
curl http://localhost:18789/health | jq .

# Check USDC balance on-chain
cast call <USDC_ADDRESS> "balanceOf(address)(uint256)" \
  0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266 --rpc-url http://localhost:8545

# Check escrow contract version
cast call <HUB_ADDRESS> "version()(string)" --rpc-url http://localhost:8545

# Restart a single agent
docker compose restart alice
```
