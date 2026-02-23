## Context

P2P Paid Value Exchange features (Payment Gate, Owner Shield, Reputation tracking, USDC Registry, ZK wiring) are fully implemented in the core layer (`internal/p2p/`), but user-facing layers lack exposure. Users have no CLI commands to check reputation or pricing, no agent tools for paid workflow orchestration, no REST endpoints for external integration, and no documentation explaining these features.

The core reputation store already tracks successes/failures/timeouts and computes trust scores, but only exposes `GetScore()` and `IsTrusted()`. The ent schema has `FirstSeen` and `LastInteraction` fields that are unused by existing API surfaces.

## Goals / Non-Goals

**Goals:**
- Expose full peer reputation details via CLI, agent tools, and REST API
- Expose pricing configuration via CLI, agent tools, and REST API
- Create skill definitions for reputation, pricing, and owner shield operations
- Update all prompts to guide agents through paid value exchange workflows
- Document Paid Value Exchange, Reputation System, and Owner Shield in feature docs
- Update example configs to demonstrate pricing and owner protection settings

**Non-Goals:**
- Changing the trust score formula or reputation calculation logic
- Adding new payment flows or modifying the Payment Gate protocol
- Implementing dynamic pricing (prices remain config-driven)
- Adding reputation history or audit trails (only current state)

## Decisions

### 1. GetDetails returns nil for unknown peers (not error)
**Rationale**: Consistent with existing `GetScore()` pattern (returns 0.0 for unknown). Callers distinguish "not found" from "error" without sentinel errors. CLI/API surfaces display "no reputation record" message.

### 2. pricingCfg stored as value on p2pComponents (not pointer)
**Rationale**: P2PPricingConfig is a small struct (bool + 2 strings + map). Value semantics are simpler and avoid nil checks. The config is read-only after initialization.

### 3. CLI reputation command uses DB directly (not ephemeral P2P node)
**Rationale**: Reputation data is in the database, not on the P2P network. No need to spin up an ephemeral node just to query local DB. The `bootLoader` provides DB access directly.

### 4. Agent tools added to existing buildP2PTools slice (not separate function)
**Rationale**: Both new tools (`p2p_price_query`, `p2p_reputation`) are P2P-scoped and share the same `p2pComponents` dependency. Keeping them in the same builder maintains consistency with existing tools.

## Risks / Trade-offs

- **[Pricing config is static]** → Prices are read from config at startup and don't change at runtime. Mitigation: acceptable for MVP; dynamic pricing can be added later via a pricing service.
- **[Reputation REST endpoint exposes trust data publicly]** → No auth on `/api/p2p/reputation`. Mitigation: follows existing pattern (status/peers/identity are also public). Only exposes aggregate scores, not sensitive data.
- **[GetDetails couples to ent schema fields]** → If ent schema changes, GetDetails must update. Mitigation: PeerDetails struct provides decoupling layer; ent fields are stable.
