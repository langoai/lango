# Design: P2P & On-Chain Examples

## Architecture

All examples follow the same Docker Compose pattern established by `examples/p2p-trading/`:

```
┌─────────────────────────────────────────────┐
│  Docker Compose (bridge network)            │
│                                             │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐    │
│  │ Agent 1 │  │ Agent 2 │  │ Agent N │    │
│  │ :1878X  │  │ :1878Y  │  │ :1878Z  │    │
│  │ P2P:900X│  │ P2P:900Y│  │ P2P:900Z│    │
│  └────┬────┘  └────┬────┘  └────┬────┘    │
│       └──────┬─────┘────────────┘          │
│              │                              │
│         ┌────▼────┐  ┌─────────┐           │
│         │  Anvil  │  │  Setup  │           │
│         │  :8545  │  │ (init)  │           │
│         └─────────┘  └─────────┘           │
└─────────────────────────────────────────────┘
```

## Key Design Decisions

### 1. Stub Contracts for Testing
Examples 2 and 5 need contract addresses for config injection but don't need real contract logic (Lango handles tool execution internally). Minimal stub contracts provide valid addresses without complex Solidity.

### 2. Payment Required for P2P Identity
P2P requires `payment.enabled: true` because DID identity is derived from the wallet key. Even P2P-only examples (1, 3) need a minimal payment config.

### 3. Polling mDNS Discovery
Fixed `sleep` for mDNS discovery is unreliable in Docker. All test scripts use a polling loop with 5s intervals and configurable timeout (60-90s).

### 4. Reputation Endpoint Requires Parameters
The `/api/p2p/reputation` endpoint requires `peer_did` query parameter. Tests verify endpoint availability via HTTP status code (400 = available) rather than response body.

## File Organization

```
examples/
├── p2p-trading/              (existing)
├── discovery-and-handshake/  (new — P2P only, 2 agents)
├── smart-account-basics/     (new — on-chain only, 1 agent + Anvil)
├── firewall-and-reputation/  (new — P2P only, 3 agents)
├── paid-tool-marketplace/    (new — P2P + on-chain, 3 agents + Anvil)
├── escrow-milestones/        (new — on-chain heavy, 2 agents + Anvil)
└── team-workspace/           (new — full stack, 4 agents + Anvil)
```
