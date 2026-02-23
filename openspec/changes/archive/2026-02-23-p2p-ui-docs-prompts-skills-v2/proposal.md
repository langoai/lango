## Why

P2P Paid Value Exchange (Payment Gate, Owner Shield, Reputation, USDC Registry, Protocol Extensions, ZK Wiring) has been fully implemented at the core level, but the user-facing layers (README, prompts, docs, skills, CLI, REST API, example configs) have not been updated to reflect these capabilities. Users cannot discover, use, or understand the paid exchange features without proper CLI commands, agent tools, documentation, and prompt guidance.

## What Changes

- Add `PeerDetails` struct and `GetDetails()` method to `internal/p2p/reputation/store.go` for full reputation info retrieval
- Add `pricingCfg` field to `p2pComponents` struct in wiring for REST API access to pricing config
- Add `lango p2p reputation` CLI subcommand to query peer trust scores and exchange history
- Add `lango p2p pricing` CLI subcommand to display pricing configuration
- Add `p2p_price_query` agent tool to query remote peer pricing before tool invocation
- Add `p2p_reputation` agent tool to check peer trust scores and exchange history
- Add `GET /api/p2p/reputation` REST endpoint for peer reputation queries
- Add `GET /api/p2p/pricing` REST endpoint for pricing configuration queries
- Create 3 new skills: `p2p-reputation`, `p2p-pricing`, `p2p-owner-shield`
- Update prompts (AGENTS.md, TOOL_USAGE.md, vault IDENTITY.md) with paid value exchange guidance
- Update docs (p2p-network.md, configuration.md, cli/p2p.md) with new sections
- Update README.md P2P section with Payment Gate, Reputation, Owner Shield
- Update example configs (alice/bob/charlie.json) with pricing, ownerProtection, minTrustScore

## Capabilities

### New Capabilities

- `p2p-reputation-cli`: CLI and REST API for querying peer reputation details and trust scores
- `p2p-pricing-cli`: CLI and REST API for querying P2P tool pricing configuration
- `p2p-value-exchange-tools`: Agent tools for price query and reputation check in paid P2P workflows

### Modified Capabilities

- `p2p-reputation`: Add `GetDetails()` method for full reputation data retrieval
- `p2p-rest-api`: Add `/reputation` and `/pricing` endpoints
- `p2p-skills`: Add 3 new skill definitions (reputation, pricing, owner-shield)
- `p2p-agent-prompts`: Update prompts with paid value exchange guidance and new tool documentation
- `p2p-documentation`: Add Paid Value Exchange, Reputation System, Owner Shield sections
- `p2p-trading-example`: Add pricing, ownerProtection, minTrustScore to example configs

## Impact

- **Core**: `internal/p2p/reputation/store.go` (new struct + method), `internal/app/wiring.go` (new field)
- **CLI**: `internal/cli/p2p/` (2 new files + p2p.go modification)
- **Agent Tools**: `internal/app/tools.go` (2 new tools in buildP2PTools)
- **REST API**: `internal/app/p2p_routes.go` (2 new handlers)
- **Skills**: `skills/p2p-reputation/`, `skills/p2p-pricing/`, `skills/p2p-owner-shield/`
- **Prompts**: `prompts/AGENTS.md`, `prompts/TOOL_USAGE.md`, `prompts/agents/vault/IDENTITY.md`
- **Docs**: `docs/features/p2p-network.md`, `docs/configuration.md`, `docs/cli/p2p.md`
- **README**: `README.md`
- **Examples**: `examples/p2p-trading/configs/{alice,bob,charlie}.json`
