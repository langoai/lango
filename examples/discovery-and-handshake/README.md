# P2P Discovery & Handshake Example

Beginner-level integration test for Lango's P2P discovery and DID-based authentication.

Spins up **2 Lango agents** (Alice, Bob) using Docker Compose, then verifies:

- mDNS peer discovery
- GossipSub agent card exchange
- DID-based handshake v1.1 (signed challenge)
- Session token establishment

No Ethereum node or payment system required — pure P2P networking.

## Architecture

```
┌──────────┐              ┌──────────┐
│  Alice    │◄────────────►│   Bob    │
│ :18789   │    mDNS +     │ :18790   │
│ P2P:9001 │   GossipSub   │ P2P:9002 │
└──────────┘              └──────────┘
```

## Configuration Highlights

| Setting | Value | Description |
|---------|-------|-------------|
| `p2p.requireSignedChallenge` | `true` | Use handshake v1.1 with ECDSA-signed challenges |
| `p2p.autoApproveKnownPeers` | `true` | Skip handshake approval for previously authenticated peers |
| `p2p.enableMdns` | `true` | Enable multicast DNS for local peer discovery |
| `p2p.gossipInterval` | `"5s"` | Interval for broadcasting agent cards via GossipSub |
| `security.interceptor.headlessAutoApprove` | `true` | Auto-approve tool invocations in headless Docker mode |

## Prerequisites

- Docker & Docker Compose v2
- `curl` — for HTTP health/API checks

No Foundry or `cast` needed — this example is pure P2P with no on-chain components.

## Quick Start

```bash
# Build the Lango Docker image and start both agents
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

| Service | Image          | Purpose                     | Port  |
|---------|----------------|-----------------------------|-------|
| `alice` | `lango:latest` | Agent 1 (research capable)  | 18789 |
| `bob`   | `lango:latest` | Agent 2 (coding capable)    | 18790 |

## Test Scenarios

1. **Health** — Both agents respond to `GET /health`
2. **P2P Status** — `GET /api/p2p/status` returns peer ID and listen addresses
3. **mDNS Discovery** — After 15s, each agent discovers the other via mDNS
4. **DID Identity** — `GET /api/p2p/identity` returns a `did:lango:` DID
5. **Gossip Agent Cards** — Each agent sees the other's capabilities via GossipSub
6. **Handshake Session** — Peers have DID info confirming completed handshake

## REST API Endpoints

| Endpoint             | Method | Description                        |
|----------------------|--------|------------------------------------|
| `/health`            | GET    | Health check                       |
| `/api/p2p/status`    | GET    | Peer ID, listen addrs, peer count  |
| `/api/p2p/peers`     | GET    | List connected peers + addresses   |
| `/api/p2p/identity`  | GET    | Local DID string                   |
| `/api/p2p/reputation`| GET    | Peer trust score and history       |

## Troubleshooting

```bash
# View all logs
make logs

# Check a specific agent
docker compose logs alice

# Manual API check
curl http://localhost:18789/api/p2p/status | jq .

# Check peer list
curl http://localhost:18789/api/p2p/peers | jq .

# Check DID identity
curl http://localhost:18790/api/p2p/identity | jq .
```
