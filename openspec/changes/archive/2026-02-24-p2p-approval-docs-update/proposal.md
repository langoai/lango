## Why

The `p2p-approval-gaps-fix` implementation added three key features — SpendingLimiter auto-approval, inbound P2P tool owner approval, and outbound payment auto-approval — but the related documentation, prompts, examples, and Makefile were not updated. Users and developers cannot discover these capabilities from docs alone.

## What Changes

- Add "Approval Pipeline" section to `docs/features/p2p-network.md` describing the 3-stage inbound gate (firewall → owner approval → execution) with Mermaid diagram
- Add "Auto-Approval for Small Amounts" subsection to the Paid Value Exchange section in P2P docs
- Add `GET /api/p2p/reputation` and `GET /api/p2p/pricing` to REST API tables and curl examples across all relevant docs
- Add `lango p2p reputation` and `lango p2p pricing` to CLI command listings
- Update `README.md` P2P feature list, config reference, and REST API section with approval pipeline details and missing config fields
- Update `prompts/TOOL_USAGE.md` with auto-approval behavior for `p2p_pay`, owner approval notes for `p2p_query`, and inbound invocation description
- Add reputation/pricing endpoint documentation to `docs/gateway/http-api.md`
- Add P2P integration note to `docs/payments/usdc.md` explaining `autoApproveBelow` cross-cutting behavior
- Add "Configuration Highlights" section to `examples/p2p-trading/README.md`
- Add `test-p2p` Makefile target for P2P and wallet spending tests

## Capabilities

### New Capabilities

- `docs-only`: Documentation-only update covering approval pipeline, auto-approval, and missing endpoint references across 7 files

### Modified Capabilities

## Impact

- **Documentation**: 6 markdown files updated (`docs/features/p2p-network.md`, `README.md`, `prompts/TOOL_USAGE.md`, `docs/gateway/http-api.md`, `docs/payments/usdc.md`, `examples/p2p-trading/README.md`)
- **Build**: `Makefile` gains `test-p2p` target
- **No code changes**: All modifications are documentation and build tooling only
