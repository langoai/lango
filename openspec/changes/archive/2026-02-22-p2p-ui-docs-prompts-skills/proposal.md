## Why

The P2P networking core (libp2p node, ZKP, handshake, firewall, discovery, protocol) is fully implemented in the core and application layers, but there is no user-facing surface: no CLI commands, no documentation, no agent prompts, and no skills reference the P2P subsystem. Users cannot interact with P2P features without this change.

## What Changes

- Add `lango p2p` CLI command group with 7 subcommands (status, peers, connect, disconnect, firewall, discover, identity) following the existing `cli/payment` bootstrap-loader pattern
- Wire `clip2p.NewP2PCmd` into `cmd/lango/main.go`
- Update agent prompts: add P2P as the 10th tool category in AGENTS.md, add P2P tool usage section in TOOL_USAGE.md, extend vault agent identity with P2P role
- Create 8 embedded skills (p2p-status, p2p-peers, p2p-connect, p2p-disconnect, p2p-discover, p2p-identity, p2p-firewall-list, p2p-firewall-add)
- Update README.md with P2P features, CLI commands, configuration reference, and architecture entry
- Update mkdocs.yml navigation with P2P feature and CLI pages
- Create new docs: features/p2p-network.md, cli/p2p.md
- Update existing docs: features/index.md (P2P card), features/a2a-protocol.md (HTTP vs P2P comparison), configuration.md (P2P config section)

## Capabilities

### New Capabilities
- `cli-p2p-management`: CLI commands for P2P node status, peer management, firewall rules, agent discovery, and identity inspection
- `p2p-agent-prompts`: Agent prompt sections describing P2P tools and vault agent P2P role
- `p2p-skills`: Embedded skill files mapping to P2P CLI commands
- `p2p-documentation`: User-facing documentation for P2P features, CLI reference, and configuration

### Modified Capabilities
- `embedded-prompt-files`: Tool category count changes from nine to ten, new P2P section added
- `mkdocs-documentation-site`: Navigation updated with P2P feature and CLI pages
- `docs-config-format`: Configuration reference expanded with P2P section

## Impact

- **CLI**: New `internal/cli/p2p/` package (8 files), `cmd/lango/main.go` import addition
- **Prompts**: 3 embedded prompt files modified (AGENTS.md, TOOL_USAGE.md, agents/vault/IDENTITY.md)
- **Skills**: 8 new skill directories under `skills/`
- **Docs**: 2 new doc files, 5 existing docs modified, mkdocs nav updated
- **Tests**: `internal/prompt/defaults_test.go` assertion updated for new tool count
- **Dependencies**: No new Go dependencies; all P2P CLI commands use existing `internal/p2p/` package
