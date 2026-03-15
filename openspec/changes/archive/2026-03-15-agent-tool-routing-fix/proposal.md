## Why

The Vault agent's embedded AGENT.md (`internal/agentregistry/defaults/vault/AGENT.md`) was missing tool name prefixes for 5 subsystems: smartaccount, economy, contract, sentinel, and escrow. When dynamic agent specs loaded from AGENT.md files, they overrode the builtin specs (from `internal/orchestration/tools.go`), causing 30+ tools with those prefixes to be unmatched to any agent. Unmatched tools fall into the orchestrator's "unmatched" bucket and are not reliably routed.

## What Changes

- Vault AGENT.md: add 15 missing prefixes (`smart_account_`, `session_key_`, `session_execute`, `policy_check`, `module_`, `spending_`, `paymaster_`, `economy_`, `escrow_`, `sentinel_`, `contract_`) and 8 missing keywords (`smart account`, `session key`, `paymaster`, `ERC-7579`, `ERC-4337`, `module`, `policy`, `deploy account`, `economy`, `budget`, `escrow`, `sentinel`, `contract`, `negotiate`, `pricing`, `risk`)
- Builtin vault spec in `tools.go`: sync prefixes and keywords to match AGENT.md, add 4 capabilityMap entries (`economy_`, `escrow_`, `sentinel_`, `contract_`)

## Capabilities

### Modified Capabilities

- `agent-routing`: Vault agent prefix and keyword lists expanded to cover smartaccount, economy, escrow, sentinel, and contract tool families

## Impact

- `internal/agentregistry/defaults/vault/AGENT.md` -- prefix/keyword sync
- `internal/orchestration/tools.go` -- builtin vault spec prefix/keyword sync + 4 capabilityMap entries
