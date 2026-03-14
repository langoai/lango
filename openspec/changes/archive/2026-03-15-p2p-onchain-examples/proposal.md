# Proposal: P2P & On-Chain Examples Expansion

## Summary

Add 6 new independent examples to the `examples/` directory, each demonstrating a distinct set of Lango platform capabilities. The examples form a progressive learning path from Beginner to Advanced.

## Motivation

The existing `examples/p2p-trading/` only covers mDNS discovery and MockUSDC transfer. Key platform features — handshake authentication, firewall ACL, reputation scoring, smart accounts, session keys, escrow milestones, team coordination, and workspace collaboration — have no runnable examples for developers to learn from.

## Scope

| # | Example | Focus | Agents | Anvil |
|---|---------|-------|--------|-------|
| 1 | `discovery-and-handshake` | P2P discovery + DID handshake v1.1 | 2 | No |
| 2 | `smart-account-basics` | ERC-4337 smart account + session keys | 1 | Yes |
| 3 | `firewall-and-reputation` | ACL rules + reputation + OwnerShield | 3 | No |
| 4 | `paid-tool-marketplace` | Pricing + prepaid/postpaid tool calls | 3 | Yes |
| 5 | `escrow-milestones` | EscrowHubV2 + milestone settlement | 2 | Yes |
| 6 | `team-workspace` | Multi-agent team + workspace + budget | 4 | Yes |

## Non-goals

- No changes to core (`internal/`) code
- No new Go code — examples are Docker Compose + shell scripts + JSON configs
- No CI/CD integration for examples (manual `make all` only)

## Success Criteria

- Each example passes all tests via `make all`
- All examples follow the established p2p-trading patterns
- Developer can run any example independently without dependencies on others
