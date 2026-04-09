# Spec: Smart Account Downstream Artifact Sync

## Purpose

Capability spec for smartaccount-downstream. See requirements below for scope and behavior contracts.

## Requirements

### REQ-1: TUI Smart Account Settings
The TUI settings editor MUST include configuration forms for all 19 SmartAccount config keys, organized into 4 categories: Smart Account (main), SA Session Keys, SA Paymaster, SA Modules.

**Scenarios:**
- Given a user opens `lango settings`, when they navigate to the Infrastructure section, then Smart Account categories are visible
- Given a user selects "Smart Account", when the form loads, then all main config fields (enabled, factory, entrypoint, safe7579, fallback, bundler) are editable
- Given a user modifies a Smart Account field and saves, then the config is persisted correctly

### REQ-2: Documentation Coverage
Feature docs, CLI docs, config docs, tool usage docs, and README MUST document all smart account capabilities matching the actual codebase.

**Scenarios:**
- Given a user reads `docs/features/smart-accounts.md`, they find architecture overview, session keys, paymaster, policy, modules, tools, and config
- Given a user reads `docs/cli/smartaccount.md`, they find all 11 CLI commands with flags and examples
- Given a user reads `docs/configuration.md`, they find all 19 SmartAccount config keys

### REQ-3: Multi-Agent Tool Routing
All 12 smart account tools MUST be routed to the vault sub-agent in multi-agent orchestration mode.

**Scenarios:**
- Given multi-agent mode is enabled and a user requests smart account operations, then the orchestrator routes to the vault agent
- Given `PartitionTools` processes smart account tools, then none fall into `Unmatched`

### REQ-4: Cross-Reference Integrity
Feature index, economy doc, and contracts doc MUST cross-reference smart accounts.

### REQ-5: Build and Deploy
Makefile MUST include `check-abi` target. Docker compose MUST include smart account env var example.
