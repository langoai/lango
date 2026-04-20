## Why

Lango's external collaboration runtime is real, but its public operator surfaces have drifted away from the actual behavior of the system. Identity, payment trust defaults, pricing control planes, and team/workspace collaboration surfaces currently force operators to reason across inconsistent CLI help, API contracts, and documentation just to understand what is actually live.

This change is needed now because the external collaboration audit is complete and has already identified the highest-impact inconsistencies blocking a coherent Phase 1/2 knowledge-exchange story and a truthful Phase 3/4 collaboration story.

## What Changes

- Align P2P identity surfaces so CLI, REST, and docs all describe the same active-DID model, including v1/v2 DID support and null/no-DID behavior.
- Canonicalize payment-side post-pay trust defaults to one inclusive threshold and update all operator-facing docs and comments to match.
- Clarify the control-plane split between provider-side P2P quote surfaces and local economy policy engines for pricing, negotiation, and escrow.
- Truth-align team, workspace, and git-bundle operator surfaces so docs and CLI no longer imply direct live control where only server-backed or tool-backed flows exist.
- Update downstream audit artifacts so they record the resolved control-plane and operator-surface decisions rather than preserving stale drift findings.

## Capabilities

### New Capabilities

- _None_

### Modified Capabilities

- `p2p-identity`: expand the identity contract to cover active DID exposure, bundle-backed v2 DID support, and no-DID/null behavior on operator surfaces.
- `p2p-rest-api`: update the identity endpoint contract so it returns the active DID when available and `null` when unavailable instead of surfacing stale or misleading semantics.
- `cli-p2p-management`: align P2P CLI help and examples with the actual guidance-oriented, server-backed, and tool-backed operator surfaces.
- `p2p-documentation`: update P2P feature and CLI docs so they truthfully describe identity, pricing, team, workspace, git, and provenance behavior.
- `p2p-pricing`: clarify that `p2p.pricing` is the provider-side public quote surface and align its threshold/default semantics.
- `p2p-settlement`: align the public settlement narrative with the authorization-driven runtime path.
- `p2p-team-payment`: unify post-pay threshold semantics and operator-facing wording for team payment coordination.
- `p2p-team-coordination`: update the documented team operator surface so runtime-only/team-runtime behavior is described honestly.
- `p2p-workspace`: update the documented workspace and shared-artifact surface so server-backed/tool-backed reality is reflected and partial chronicler wiring is acknowledged.

## Impact

- Affected code: `internal/cli/p2p/*`, `internal/app/p2p_routes.go`, `internal/p2p/paygate/*`, `internal/p2p/team/*`, `internal/p2p/trustpolicy/*`
- Affected docs: `docs/features/p2p-network.md`, `docs/features/economy.md`, `docs/cli/p2p.md`, `docs/gateway/http-api.md`, `docs/architecture/external-collaboration-audit.md`
- Affected APIs: `GET /api/p2p/identity`, `GET /api/p2p/pricing`
- Affected systems: external exchange operator model, payment trust policy, team/workspace operator surfaces, audit closeout guidance
