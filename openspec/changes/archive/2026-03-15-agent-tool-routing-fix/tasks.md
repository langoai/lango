## 1. Vault AGENT.md Prefix Sync

- [x] 1.1 Add `smart_account_`, `session_key_`, `session_execute`, `policy_check`, `module_`, `spending_`, `paymaster_` prefixes
- [x] 1.2 Add `economy_`, `escrow_`, `sentinel_`, `contract_` prefixes
- [x] 1.3 Add keywords: `smart account`, `session key`, `paymaster`, `ERC-7579`, `ERC-4337`, `module`, `policy`, `deploy account`
- [x] 1.4 Add keywords: `economy`, `budget`, `escrow`, `sentinel`, `contract`, `negotiate`, `pricing`, `risk`

## 2. Builtin Vault Spec Sync (tools.go)

- [x] 2.1 Sync vault Prefixes slice with AGENT.md (15 total prefixes)
- [x] 2.2 Sync vault Keywords slice with AGENT.md (31 total keywords)
- [x] 2.3 Add capabilityMap entries: `economy_`, `escrow_`, `sentinel_`, `contract_`

## 3. Verification

- [x] 3.1 go build ./... passes
- [x] 3.2 go test ./... passes
