# Proposal: Downstream Smart Account Artifact Sync

## Problem
Four commits on `feature/p2p-smart-account` added 51 new internal source files and 8 CLI command files for ERC-7579 smart accounts, but downstream artifacts (TUI settings, documentation, README, Makefile, Docker, prompts, multi-agent routing) were not updated.

## Solution
Sync all downstream artifacts to reflect the smart account subsystem:

1. **TUI Settings** — Add Smart Account configuration forms (4 categories: main, session, paymaster, modules)
2. **README.md** — Add smart account features and CLI commands
3. **Feature Documentation** — Create `docs/features/smart-accounts.md`
4. **CLI Documentation** — Create `docs/cli/smartaccount.md`, update `docs/cli/index.md`
5. **Configuration Documentation** — Add SmartAccount section to `docs/configuration.md`
6. **Tool Usage Documentation** — Add 12 smart account tools to `prompts/TOOL_USAGE.md`
7. **Cross-References** — Update feature index, economy doc, contracts doc
8. **Build/Deploy** — Add `check-abi` Makefile target, Docker env var
9. **Multi-Agent Routing** — Add smart account tool prefixes to vault agent, update capability map, update agent identity

## Scope
- 9 work units, all independent
- Code changes in TUI settings (4 files) and orchestration routing (3 files)
- Documentation changes across 11 files
- No changes to core smart account logic
