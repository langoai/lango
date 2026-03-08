# Design: Smart Account Downstream Sync

## Approach
All 9 work units are independent and can be parallelized. Each WU touches non-overlapping files.

## Key Decisions

### TUI Forms (WU-1)
- Follow existing form patterns (e.g., `forms_economy.go`, `forms_p2p.go`)
- 4 separate form constructors for better organization
- Config key prefix: `sa_` to avoid collisions with existing keys
- Use `InputSelect` for provider and fallback mode enums

### Multi-Agent Routing (WU-9 extension)
- Add 7 tool prefixes to vault agent: `smart_account_`, `session_key_`, `session_execute`, `policy_check`, `module_`, `spending_`, `paymaster_`
- Add corresponding entries to `capabilityMap` for auto-generated capability descriptions
- Update vault agent instruction text to mention smart account operations
- Update vault IDENTITY.md prompt file

### Documentation Strategy
- Feature doc (`smart-accounts.md`): comprehensive, based on actual source code analysis
- CLI doc (`smartaccount.md`): all 11 commands with actual flags and output format
- Config doc: all 19 keys with types, defaults, descriptions
- Tool usage: all 12 agent tools with parameters, safety levels, workflows

## File Impact

| Layer | Files Changed | Files Created |
|-------|--------------|---------------|
| TUI | menu.go, editor.go, state_update.go | forms_smartaccount.go |
| Orchestration | tools.go | — |
| Prompts | TOOL_USAGE.md, AGENTS.md, vault/IDENTITY.md | — |
| Docs | index.md, economy.md, contracts.md, configuration.md, cli/index.md | smart-accounts.md, cli/smartaccount.md |
| Build | Makefile, docker-compose.yml | — |
| README | README.md | — |
