## Purpose

Capability spec for p2p-onchain-examples. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: p2p-onchain-examples capability documented
The p2p-onchain-examples capability SHALL be documented through the sections in this spec. This requirement is a structural placeholder that satisfies the canonical openspec format; detailed behavior contracts live in the descriptive sections of this file.

#### Scenario: Spec file is readable
- **WHEN** the p2p-onchain-examples spec.md file is read
- **THEN** it SHALL describe the capability's behavior in sections below

# Spec: P2P & On-Chain Examples

## Overview

Seven Docker Compose-based integration examples demonstrating Lango's P2P networking, on-chain economy, and multi-agent coordination features.

## Shared Infrastructure

### File Structure Pattern
Each example follows the same directory structure:
- `README.md` — Architecture diagram, config highlights, test scenarios, troubleshooting
- `Makefile` — Targets: build, up, test, down, clean, logs, all
- `docker-compose.yml` — Services on isolated bridge network
- `configs/*.json` — Per-agent JSON config files
- `secrets/*-passphrase.txt` — Test passphrases (never for production)
- `docker-entrypoint-*.sh` — Bootstrap: keyfile → config import → wallet key → serve
- `scripts/wait-for-health.sh` — Poll URL until HTTP 200 (2s interval, configurable timeout)
- `scripts/test-*.sh` — Colored PASS/FAIL output, section headers, counter summary

### Test Script Pattern
- Colors: RED (fail), GREEN (pass), YELLOW (section headers)
- Functions: `pass()`, `fail()`, `section()`
- mDNS discovery: most examples use polling loops (5s interval, up to 60-90s), while `p2p-trading` currently uses a fixed `sleep 15` warm-up before peer checks
- Reputation endpoint: check HTTP status code (400 = available, requires peer_did param)
- Exit 0 on all pass, exit 1 on any failure

### Config Requirements
- P2P requires `payment.enabled: true` (wallet needed for DID identity derivation)
- All agents need `AGENT_PRIVATE_KEY` env var for wallet key injection
- Anvil deterministic keys: accounts[0-3] for agents, account[9] for deployer

## Example Specifications

### 1. discovery-and-handshake (Beginner)
- **Agents**: Alice (18789, P2P:9001), Bob (18790, P2P:9002)
- **Features**: mDNS, gossip cards, signed challenge handshake v1.1, session tokens
- **Config**: `requireSignedChallenge: true`, no pricing, open firewall
- **Checks include**: health, P2P status, mDNS discovery, DID identity, gossip, handshake

### 2. smart-account-basics (Beginner/Intermediate)
- **Agents**: 1 agent (18789) + Anvil
- **Features**: Smart account deploy, session keys, policy engine, spending hook
- **Contracts**: MockUSDC, EntryPointStub, FactoryStub
- **Config**: `smartAccount.enabled: true`, P2P disabled
- **Checks include**: health, contract deployment, SA deploy/info, session keys, policy, spending

### 3. firewall-and-reputation (Intermediate)
- **Agents**: Alice (provider, 18789), Bob (trusted, 18790), Charlie (untrusted, 18791)
- **Features**: ACL rules (allow/deny per tool), rate limiting, reputation, OwnerShield PII
- **Config**: Alice has restrictive firewall (allow knowledge_search/web_search, rate 5; deny browser_navigate/file_read/shell_exec), `minTrustScore: 0.5`, PII in ownerProtection
- **Checks include**: health, discovery, firewall, DID, reputation, PII shield, trust config, pricing

### 4. p2p-trading (Intermediate)
- **Agents**: Alice (18789), Bob (18790), Charlie (18791) + Anvil
- **Features**: mDNS discovery, DID identity, public pricing, and direct on-chain USDC transfer between peers
- **Config**: `p2p.pricing.enabled: true`, `p2p.pricing.perQuery: "0.10"`, `p2p.autoApproveKnownPeers: true`, `payment.limits.autoApproveBelow: "50.00"`
- **Checks include**: health, P2P status, peer discovery, DID identity, on-chain balances, payment transfer

### 5. paid-tool-marketplace (Intermediate/Advanced)
- **Agents**: Alice (seller, 18789), Bob (buyer, 18790), Charlie (high-trust, 18791) + Anvil
- **Features**: Tool pricing, prepaid invoke, postpaid invoke, on-chain settlement
- **Config**: Alice has `pricing.toolPrices` (4 tools), `trustThresholds.postPayMinScore: 0.8`
- **Checks include**: health, discovery, USDC balances, pricing, DID, reputation, prepay transfer, postpay settlement

### 6. escrow-milestones (Advanced)
- **Agents**: Alice (buyer, 18789), Bob (seller, 18790) + Anvil
- **Features**: EscrowHubV2, MilestoneSettler, budget, risk assessment, milestone release
- **Contracts**: MockUSDC, EscrowHubV2Stub, MilestoneSettlerStub, DirectSettlerStub
- **Config**: `economy.enabled: true`, `escrow.enabled: true`, `budget.defaultMax: "50.00"`, `risk.escrowThreshold: "5.00"`
- **Checks include**: health, contracts, discovery, DID, balances, escrow verification, on-chain simulation, budget, economy config

### 7. team-workspace (Advanced)
- **Agents**: Leader (18789), Worker1 (18790), Worker2 (18791), Worker3 (18792) + Anvil
- **Features**: Team formation, task delegation, health monitoring, workspace, contribution tracking, budget
- **Config**: Leader has `workspace.enabled: true`, `team.healthCheckInterval: "15s"`, `economy.budget.defaultMax: "100.00"`
- **Checks include**: health, discovery, DID, P2P status, USDC balances, team config, capabilities, reputation, budget simulation, health monitoring
