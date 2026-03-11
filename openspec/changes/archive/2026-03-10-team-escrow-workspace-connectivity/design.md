# Design: P2P Team-Escrow-Workspace Connectivity

## Architecture

All connectivity is achieved through the existing EventBus pattern — no new cross-package imports needed between Core packages. Bridges live in the `internal/app/` layer which can import everything.

```
┌─────────────┐    EventBus    ┌──────────────────┐
│   Team      │───────────────►│ Team-Escrow      │
│ Coordinator │    events      │ Bridge           │
│             │                │ (app layer)      │
│  FormTeam   │                │                  │
│  Delegate   │                │  → Create Escrow │
│  Disband    │                │  → Complete MS   │
│             │                │  → Release/Refund│
└─────────────┘                └──────────────────┘
      │                               │
      │ events                        │ calls
      ▼                               ▼
┌─────────────┐                ┌──────────────────┐
│ Team-Budget │                │  Escrow Engine   │
│ Bridge      │                │                  │
│             │                │  → Create        │
│ → Allocate  │                │  → Fund          │
│ → Reserve   │                │  → Activate      │
│ → Record    │                │  → Complete MS   │
└─────────────┘                │  → Release       │
      │                        │  → Refund        │
      │ calls                  └──────────────────┘
      ▼
┌─────────────┐                ┌──────────────────┐
│   Budget    │                │ Workspace-Team   │
│   Engine    │                │ Bridge           │
│             │                │                  │
│  → Allocate │                │ → Create WS      │
│  → Reserve  │                │ → Track Contrib   │
│  → Record   │                │ → Cleanup         │
└─────────────┘                └──────────────────┘
```

## Key Decisions

### Event-Driven Bridges (not direct calls)
Bridges subscribe to EventBus events rather than being called directly. This preserves the existing loose coupling pattern and avoids import cycles.

### sync.Map for Cross-Event State
Each bridge uses `sync.Map` for teamID→escrowID / teamID→workspaceID mapping. No struct needed — closures capture the map.

### Callback Pattern for Protocol Handler
TeamHandler is a function type (like NegotiateHandler), avoiding import of team package from protocol package.

### TeamRouter for Type-Safe Dispatch
JSON marshal/unmarshal is used to convert `map[string]interface{}` payloads to typed structs. This follows the NegotiatePayload pattern.

### USDC Amount Conversion
Float64 → big.Int with 6 decimal places (USDC standard). Helper functions `floatToUSDC` and `floatToBudgetAmount` are kept separate to avoid coupling.

## Files

| File | Type | Purpose |
|------|------|---------|
| `internal/p2p/team/coordinator.go` | Modified | Publish 5 missing events |
| `internal/p2p/team/coordinator_test.go` | Modified | Event publishing tests |
| `internal/p2p/protocol/handler.go` | Modified | TeamHandler type + switch cases |
| `internal/p2p/protocol/team_handler.go` | New | TeamRouter dispatch logic |
| `internal/p2p/protocol/remote_agent.go` | Modified | 3 team methods |
| `internal/app/tools_team.go` | New | 5 team agent tools |
| `internal/app/tools_team_escrow.go` | New | 2 convenience tools |
| `internal/app/bridge_team_escrow.go` | New | Event-driven escrow bridge |
| `internal/app/bridge_team_budget.go` | New | Event-driven budget bridge |
| `internal/app/bridge_workspace_team.go` | New | Event-driven workspace bridge |
| `internal/app/bridge_integration_test.go` | New | 5 integration tests |
| `internal/app/app.go` | Modified | Wire bridges + tools |
| `internal/app/wiring_p2p.go` | Modified | Wire team protocol handler |
