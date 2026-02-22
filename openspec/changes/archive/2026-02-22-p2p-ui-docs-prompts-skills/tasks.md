## 1. CLI P2P Command Group

- [x] 1.1 Create `internal/cli/p2p/p2p.go` with `NewP2PCmd`, `p2pDeps` struct, and `initP2PDeps` using bootstrap Result loader pattern
- [x] 1.2 Create `internal/cli/p2p/status.go` with `lango p2p status [--json]` command
- [x] 1.3 Create `internal/cli/p2p/peers.go` with `lango p2p peers [--json]` command using tabwriter
- [x] 1.4 Create `internal/cli/p2p/connect.go` with `lango p2p connect <multiaddr>` command
- [x] 1.5 Create `internal/cli/p2p/disconnect.go` with `lango p2p disconnect <peer-id>` command
- [x] 1.6 Create `internal/cli/p2p/firewall.go` with `lango p2p firewall [list|add|remove]` subcommands
- [x] 1.7 Create `internal/cli/p2p/discover.go` with `lango p2p discover [--tag] [--json]` command
- [x] 1.8 Create `internal/cli/p2p/identity.go` with `lango p2p identity [--json]` command
- [x] 1.9 Wire `clip2p.NewP2PCmd` into `cmd/lango/main.go` with bootstrap loader

## 2. Agent Prompts

- [x] 2.1 Update `prompts/AGENTS.md`: change "nine" to "ten" tool categories, add P2P Network bullet
- [x] 2.2 Update `prompts/TOOL_USAGE.md`: add P2P Networking Tool section with all P2P tool guidelines
- [x] 2.3 Update `prompts/agents/vault/IDENTITY.md`: add P2P peer management to vault agent role

## 3. Embedded Skills

- [x] 3.1 Create `skills/p2p-status/SKILL.md` (type: script, `lango p2p status`)
- [x] 3.2 Create `skills/p2p-peers/SKILL.md` (type: script, `lango p2p peers`)
- [x] 3.3 Create `skills/p2p-connect/SKILL.md` (type: script, `lango p2p connect $MULTIADDR`)
- [x] 3.4 Create `skills/p2p-disconnect/SKILL.md` (type: script, `lango p2p disconnect $PEER_ID`)
- [x] 3.5 Create `skills/p2p-discover/SKILL.md` (type: script, `lango p2p discover`)
- [x] 3.6 Create `skills/p2p-identity/SKILL.md` (type: script, `lango p2p identity`)
- [x] 3.7 Create `skills/p2p-firewall-list/SKILL.md` (type: script, `lango p2p firewall list`)
- [x] 3.8 Create `skills/p2p-firewall-add/SKILL.md` (type: script, `lango p2p firewall add`)

## 4. Documentation

- [x] 4.1 Create `docs/features/p2p-network.md` with overview, identity, handshake, firewall, discovery, ZK circuits, config, CLI sections
- [x] 4.2 Create `docs/cli/p2p.md` with usage, flags, examples for all P2P commands
- [x] 4.3 Update `docs/features/index.md`: add P2P Network card and Feature Status row
- [x] 4.4 Update `docs/features/a2a-protocol.md`: add A2A-over-HTTP vs A2A-over-P2P comparison
- [x] 4.5 Update `mkdocs.yml`: add P2P feature and CLI pages to navigation
- [x] 4.6 Update `docs/configuration.md`: add P2P Network config section with JSON example and table
- [x] 4.7 Update `README.md`: add P2P to features, CLI commands, config table, architecture tree, and new P2P section

## 5. Test Updates

- [x] 5.1 Update `internal/prompt/defaults_test.go`: change "nine tool categories" assertion to "ten tool categories"
- [x] 5.2 Verify `go build ./...` passes
- [x] 5.3 Verify `go test ./...` passes
