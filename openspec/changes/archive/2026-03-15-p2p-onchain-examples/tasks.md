# Tasks: P2P & On-Chain Examples

## Implementation

- [x] Create `examples/discovery-and-handshake/` (10 files)
- [x] Create `examples/smart-account-basics/` (12 files)
- [x] Create `examples/firewall-and-reputation/` (12 files)
- [x] Create `examples/paid-tool-marketplace/` (14 files)
- [x] Create `examples/escrow-milestones/` (15 files)
- [x] Create `examples/team-workspace/` (16 files)

## Bug Fixes During Verification

- [x] Add `payment.enabled` to P2P-only examples (1, 3) — P2P requires wallet for DID
- [x] Add `AGENT_PRIVATE_KEY` env vars to P2P-only docker-compose files
- [x] Update P2P-only entrypoints to inject wallet keys
- [x] Replace fixed `sleep 15` with polling loop for mDNS discovery (all examples)
- [x] Fix reputation endpoint test — requires `peer_did` param, check HTTP 400 instead
- [x] Fix `cast send` in escrow test — use `docker compose exec -T anvil` wrapper

## Verification

- [x] Example 1: discovery-and-handshake — 12/12 tests passed
- [x] Example 2: smart-account-basics — 9/9 tests passed
- [x] Example 3: firewall-and-reputation — 19/19 tests passed
- [x] Example 4: paid-tool-marketplace — 21/21 tests passed
- [x] Example 5: escrow-milestones — 22/22 tests passed
- [x] Example 6: team-workspace — 34/34 tests passed
