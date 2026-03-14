# Team Workspace — Multi-Agent Team + Collaborative Workspace

An advanced integration example demonstrating **multi-agent team formation**, **task delegation**, **health monitoring**, **workspace collaboration**, **contribution tracking**, and **team budget management** using 4 Lango agents (1 Leader + 3 Workers) on a local Anvil blockchain.

## Architecture

```
                    +-------------------+
                    |      Anvil        |
                    | (Local Blockchain)|
                    |    :8545          |
                    +--------+----------+
                             |
              +--------------+--------------+
              |              |              |
     +--------+--+   +------+----+  +------+----+
     |  Leader   |   |  Worker1  |  |  Worker2  |
     |  :18789   |   |  :18790   |  |  :18791   |
     |  mgmt     |   |  research |  |  coding   |
     +-----+-----+   +-----+-----+  +-----+-----+
           |               |              |
           |         +-----+-----+        |
           +---------+  Worker3  +--------+
                     |  :18792   |
                     | research  |
                     | + coding  |
                     +-----------+

     P2P mesh via mDNS discovery (libp2p)
     USDC payments via MockUSDC (ERC-20)
```

## Features

| Feature | Description |
|---|---|
| **Team Formation** | Leader discovers workers via mDNS and forms a P2P team |
| **Task Delegation** | Leader assigns tasks based on agent capabilities |
| **Health Monitoring** | Periodic heartbeat checks with configurable intervals |
| **Workspace Collaboration** | Shared workspace with contribution tracking |
| **Git State Sync** | Track git state across team members |
| **Contribution Tracking** | Record and attribute agent contributions |
| **Team Budget** | Economy-enabled budget with USDC milestone payments |
| **Escrow** | Milestone-based escrow for task payments |
| **Reputation** | Trust scores influence team membership eligibility |
| **Signed Challenges** | P2P handshakes use signed challenge protocol |

## Configuration Highlights

### Team Settings
```json
{
  "p2p.team.healthCheckInterval": "15s",
  "p2p.team.maxMissedHeartbeats": 3,
  "p2p.team.minReputationScore": 0.3,
  "p2p.team.gitStateTracking": true
}
```

### Workspace Settings
```json
{
  "p2p.workspace.enabled": true,
  "p2p.workspace.maxWorkspaces": 5,
  "p2p.workspace.contributionTracking": true
}
```

### Economy Settings
```json
{
  "economy.enabled": true,
  "economy.budget.defaultMax": "100.00",
  "economy.escrow.enabled": true
}
```

## Prerequisites

- Docker and Docker Compose
- Go 1.25+ (to build the `lango` image)
- `curl` and `jq` (for test scripts)

## Quick Start

```bash
# Build the Lango Docker image and start all services
make all

# Or step by step:
make build   # Build Docker image
make up      # Start all containers
make test    # Run integration tests
make down    # Stop and remove containers
```

## Services

| Service | Port | Description |
|---|---|---|
| `anvil` | 8545 | Local Ethereum blockchain (Foundry) |
| `setup` | — | Deploys MockUSDC, mints tokens, then exits |
| `leader` | 18789 | Team leader — orchestrates team and manages budget |
| `worker1` | 18790 | Worker 1 — research specialist |
| `worker2` | 18791 | Worker 2 — coding specialist |
| `worker3` | 18792 | Worker 3 — research + coding generalist |

## Test Scenarios

The integration test suite (`scripts/test-team.sh`) covers 10 sections:

1. **Health Checks** — All 4 agents respond with `{"status":"ok"}`
2. **P2P Discovery** — Each agent discovers 3 peers via mDNS
3. **DID Identity** — Each agent has a valid `did:lango:` identity
4. **P2P Status** — Each agent reports P2P status with peerId
5. **USDC Balances** — Each agent holds 1000.00 USDC after mint
6. **Team Configuration** — Leader has correct team/workspace/economy config
7. **Agent Capabilities** — Leader sees peers with capability metadata
8. **Reputation Baseline** — All agents have reputation endpoints available
9. **Team Budget Simulation** — Leader pays Worker1 a 10 USDC milestone
10. **Worker Health Monitoring** — All workers remain healthy under monitoring

## Anvil Test Accounts

| Role | Address | Private Key |
|---|---|---|
| Deployer | `0xa0Ee7A142d267C1f36714E4a8F75612F20a79720` | `0x2a871d...` |
| Leader | `0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266` | `0xac0974...` |
| Worker1 | `0x70997970C51812dc3A010C7d01b50e0d17dc79C8` | `0x59c699...` |
| Worker2 | `0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC` | `0x5de411...` |
| Worker3 | `0x90F79bf6EB2c4f870365E785982E1f101E93b906` | `0x47e179...` |

## Troubleshooting

### Agents not discovering peers
- mDNS discovery takes ~20 seconds; the test script waits accordingly
- Ensure all containers are on the same Docker network (`team-net`)
- Check logs: `docker compose logs leader`

### USDC balance is zero
- Verify the `setup` container completed successfully: `docker compose logs setup`
- Check that Anvil is healthy: `curl http://localhost:8545`

### Health check timeouts
- Default timeout is 120 seconds; agents need time to bootstrap
- Check individual agent logs: `docker compose logs worker1`

### Build failures
- Ensure you are in the `examples/team-workspace/` directory
- The Dockerfile is at the repository root (`../../Dockerfile`)

### Port conflicts
- Leader: 18789, Worker1: 18790, Worker2: 18791, Worker3: 18792
- Anvil: 8545
- If these ports are in use, stop conflicting services first
