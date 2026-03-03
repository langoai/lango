## Context

The P2P payment system currently treats all peers identically: every paid tool invocation requires upfront EIP-3009 authorization regardless of the peer's history. The `Gate.SubmitOnChain()` method is a TODO placeholder that never submits real transactions.

The reputation system (`internal/p2p/reputation/`) already tracks per-peer trust scores based on exchange outcomes. The event bus (`internal/eventbus/`) provides decoupled pub/sub. These existing foundations enable trust-based routing and event-driven settlement.

## Goals / Non-Goals

**Goals:**
- Route payment flow by trust: post-pay for trusted peers, prepay for medium-trust, blocked for low-trust
- Implement real on-chain settlement via `transferWithAuthorization` submission
- Create feedback loop: settlement outcome updates peer reputation
- Maintain backward compatibility for existing prepay flow

**Non-Goals:**
- Persistent deferred ledger (in-memory is sufficient for MVP; Ent table is a follow-up)
- Gas cost optimization or meta-transactions
- Multi-chain settlement
- Dispute resolution for failed post-pay settlements

## Decisions

### D1: Trust threshold placement — in payment gate, not firewall
**Decision**: Post-pay routing lives in `paygate.Gate.Check()`, not in the firewall.
**Rationale**: The firewall handles access control (allow/deny). Payment tier routing is a business decision about _how_ to charge, not _whether_ to allow. Keeping them separate maintains the single-responsibility boundary.

### D2: Event-driven settlement over inline submission
**Decision**: Settlement happens asynchronously via `eventbus.ToolExecutionPaidEvent` → `settlement.Service`, not inline in the handler.
**Rationale**: On-chain submission can take seconds to minutes (receipt confirmation). Blocking the P2P response stream would cause timeouts. Event-driven processing decouples tool execution latency from settlement latency. Also allows retry without re-executing the tool.

### D3: Direct eventbus import in protocol handler
**Decision**: `protocol.Handler` imports `internal/eventbus` directly instead of using a callback interface.
**Rationale**: The eventbus package depends only on `sync` — no import cycle risk. Using a direct import is simpler than the callback/adapter pattern and the Bus is already designed as a dependency-free package.

### D4: Nonce serialization via mutex
**Decision**: `settlement.Service.nonceMu` serializes `buildAndSignTx()` calls.
**Rationale**: Concurrent `PendingNonceAt` calls can return the same nonce, causing one transaction to be rejected. A mutex is the simplest correct solution. At the expected throughput (~1 settlement/sec), contention is negligible.

### D5: In-memory deferred ledger
**Decision**: `DeferredLedger` is in-memory (`map[string]*DeferredEntry`), not Ent-backed.
**Rationale**: Post-pay obligations are short-lived (settled within seconds of tool execution). Persistence adds complexity for minimal benefit at this stage. The ledger is primarily for observability. Can upgrade to Ent table if durability becomes a requirement.

## Risks / Trade-offs

- **[Volatile ledger]** → Restart loses unsettled post-pay records → Mitigated by short settlement window; upgrade to Ent later if needed
- **[Gas cost on seller]** → Seller pays gas for `transferWithAuthorization` tx → Tool prices should include gas margin; future: gas abstraction
- **[Nonce serialization bottleneck]** → Sequential tx building limits throughput → Acceptable at current scale; can implement nonce pool if needed
- **[Receipt timeout]** → Network congestion could exceed 2-minute confirmation window → Configurable timeout; tx is still on-chain even if confirmation times out
