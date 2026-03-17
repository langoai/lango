# Firewall & Reputation — ACL Rules + Trust Scoring

Intermediate integration example demonstrating Lango's firewall access control, per-peer rate limiting, reputation-based trust scoring, and OwnerShield PII protection.

Spins up **3 Lango agents** (Alice, Bob, Charlie) using Docker Compose — no Anvil or payment system required:

- **Alice** — Provider with **restrictive** firewall rules (allows only `knowledge_search`, `web_search`; denies `browser_navigate`, `file_read`, `shell_exec`), OwnerShield PII protection, and a higher `minTrustScore` of 0.5
- **Bob** — Trusted client with open firewall rules and default trust threshold (0.3)
- **Charlie** — Untrusted client with open firewall rules and default trust threshold (0.3)

## Architecture

```
┌──────────────────┐     ┌──────────────────┐     ┌──────────────────┐
│  Alice (Provider) │◄───►│   Bob (Trusted)   │◄───►│Charlie (Untrusted)│
│     :18789       │     │     :18790        │     │     :18791        │
│   P2P:9001       │     │   P2P:9002        │     │   P2P:9003        │
│                  │     │                   │     │                   │
│ Firewall: strict │     │ Firewall: open    │     │ Firewall: open    │
│ Trust: >= 0.5    │     │ Trust: >= 0.3     │     │ Trust: >= 0.3     │
│ OwnerShield: ON  │     │                   │     │                   │
└──────────────────┘     └───────────────────┘     └───────────────────┘
```

## Configuration Highlights

| Setting | Alice | Bob / Charlie | Description |
|---------|-------|---------------|-------------|
| `p2p.firewallRules` | allow `knowledge_search`, `web_search` (rate 5); deny `browser_navigate`, `file_read`, `shell_exec` | allow `*` | ACL-based tool access control |
| `p2p.minTrustScore` | `0.5` | `0.3` | Minimum reputation to interact |
| `p2p.ownerProtection.ownerName` | `"Alice Smith"` | agent name | PII to protect |
| `p2p.ownerProtection.ownerEmail` | `"alice@example.com"` | `""` | Email PII |
| `p2p.ownerProtection.ownerPhone` | `"+1-555-0100"` | `""` | Phone PII |
| `p2p.ownerProtection.blockConversations` | `true` | `true` | Block PII leakage in conversations |
| `security.interceptor.headlessAutoApprove` | `true` | `true` | Auto-approve in headless Docker mode |

## Prerequisites

- Docker & Docker Compose v2
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

| Service   | Image          | Purpose                                    | Port  |
|-----------|----------------|--------------------------------------------|-------|
| `alice`   | `lango:latest` | Provider with restrictive firewall + PII   | 18789 |
| `bob`     | `lango:latest` | Trusted client                             | 18790 |
| `charlie` | `lango:latest` | Untrusted client                           | 18791 |

## Test Scenarios

1. **Health Checks** — All 3 agents respond to `GET /health`
2. **P2P Discovery** — After 15s mDNS, each agent sees >= 2 peers
3. **Firewall Configuration** — Alice has active P2P with firewall rules
4. **DID Identity** — Each agent has a `did:lango:` DID
5. **Reputation Scores** — Reputation endpoint is available on all agents
6. **Owner Protection (PII Shield)** — Alice's PII (email, phone) is not leaked in identity responses
7. **Trust Score Configuration** — Alice requires 0.5, Bob/Charlie use 0.3
8. **Pricing Configuration** — Pricing endpoint availability check

## REST API Endpoints

| Endpoint             | Method | Description                        |
|----------------------|--------|------------------------------------|
| `/health`            | GET    | Health check                       |
| `/api/p2p/status`    | GET    | Peer ID, listen addrs, peer count  |
| `/api/p2p/peers`     | GET    | List connected peers + addresses   |
| `/api/p2p/identity`  | GET    | Local DID string                   |
| `/api/p2p/reputation`| GET    | Peer trust scores and history      |
| `/api/p2p/pricing`   | GET    | Tool pricing configuration         |

## Troubleshooting

```bash
# View all logs
make logs

# Check a specific agent
docker compose logs alice

# Manual API check
curl http://localhost:18789/api/p2p/status | jq .

# Check firewall rules via identity
curl http://localhost:18789/api/p2p/identity | jq .

# Check reputation
curl http://localhost:18789/api/p2p/reputation | jq .
```
