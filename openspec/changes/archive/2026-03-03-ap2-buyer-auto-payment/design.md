## Context

The P2P payment system has a complete seller-side flow (paygate → trust-based routing → settlement) and X402 HTTP 402 auto-payment. The buyer side has `p2p_price_query` for price discovery and `p2p_pay` for direct payments, but no automated tool that combines price checking, authorization signing, spending enforcement, and paid tool invocation into a single call. The existing `InvokeToolPaid()` API in `remote_agent.go` and `eip3009.Sign()` are ready but not exposed through any agent tool.

## Goals / Non-Goals

**Goals:**
- Provide a single `p2p_invoke_paid` tool that handles the entire buyer-side paid invocation flow
- Reuse existing infrastructure: `eip3009`, `SpendingLimiter`, `InvokeToolPaid()`, `QueryPrice()`
- Ensure the buyer's `authToMap()` output is wire-compatible with seller's `parseAuthorization()`
- Connect the `SpendingLimiter` to P2P purchases (previously X402-only)

**Non-Goals:**
- Changing the seller-side paygate or settlement logic
- Adding new spending limit configuration (reuses existing `payment.spending.*` config)
- Multi-step approval UI for amounts exceeding auto-approve threshold (returns status for agent to handle)

## Decisions

### 1. Single tool vs. multi-step tool chain
**Decision**: Single `p2p_invoke_paid` tool that handles the full flow internally.
**Rationale**: The agent should be able to invoke a paid remote tool in one step. Breaking it into multiple tools (query → approve → sign → invoke) adds unnecessary complexity and round-trips. The tool handles free tools transparently by routing to `InvokeTool()` when `quote.IsFree` is true.
**Alternative**: Separate `p2p_authorize` + `p2p_invoke_with_auth` tools — rejected because it requires the agent to orchestrate multi-step payment flows.

### 2. USDC address lookup at build time vs. call time
**Decision**: Look up `contracts.LookupUSDC(chainID)` at tool build time in `buildP2PPaidInvokeTool()`.
**Rationale**: Chain ID is fixed at startup and won't change during runtime. Early lookup means the tool is not registered if the chain is unsupported, failing fast.
**Alternative**: Lookup per invocation — rejected because it adds unnecessary per-call overhead for a value that never changes.

### 3. Auto-approval gate
**Decision**: Use `SpendingLimiter.IsAutoApprovable()` to gate automatic payment. If not auto-approvable, return an `approval_required` status instead of failing.
**Rationale**: This allows the agent (or user) to decide how to handle amounts above the auto-approve threshold without hard-failing. The tool remains safe by never spending more than the configured threshold automatically.

### 4. Authorization serialization format
**Decision**: `authToMap()` produces the exact field names and types that `paygate.parseAuthorization()` expects: hex addresses, decimal string big ints, hex `[32]byte`, and `float64` for `v`.
**Rationale**: Wire compatibility is critical. The seller's parser uses `getHexAddress`, `getBigInt` (string path), `getBytes32` (hex), and `getUint8` (float64). Our output matches all these expectations exactly.

## Risks / Trade-offs

- **[Race between price query and signing]** → The price quote has an expiry (`QuoteExpiry`). The 10-minute `paidInvokeDefaultDeadline` on the EIP-3009 authorization is well within typical quote windows (5 minutes). If the quote expires between query and invoke, the seller returns `payment_required` and the agent can retry.
- **[SpendingLimiter shared with X402]** → Both X402 and P2P purchases draw from the same daily budget. This is intentional — a single spending boundary prevents runaway costs from either channel.
- **[No retry on payment rejection]** → The tool returns the rejection status rather than retrying. The agent layer can decide to retry with updated parameters.
