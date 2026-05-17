# CLI Reference

Lango provides a comprehensive command-line interface built with [Cobra](https://github.com/spf13/cobra). Every command supports `--help` for detailed usage information.

## Quick Reference

| Command | Description |
|---------|-------------|
| `lango` | Launch standalone mission workbench TUI (default entry point) |
| `lango cockpit` | Launch explicit multi-panel operator dashboard with Mission Control, chat transcript visibility, and operator detail pages |
| `lango chat` | Launch focused chat TUI with tool lifecycle visibility and inline approval controls |
| `lango serve` | Start the gateway server |
| `lango version` | Print version and build info |
| `lango health` | Check gateway health |
| `lango status` | [Unified status dashboard](status.md) (health, config, features) |
| `lango status dead-letter-summary` | Show overview counts, grouped reason/actor/dispatch-family buckets, configurable raw top-N sections, and recent trend windows for the current dead-letter backlog |
| `lango status dead-letters` | List current dead-lettered post-adjudication executions with latest-family and any-match-family filtering |
| `lango status dead-letter <transaction-receipt-id>` | Show detailed dead-letter execution status for one transaction |
| `lango status dead-letter retry <transaction-receipt-id>` | Request retry for one dead-lettered post-adjudication execution with structured follow-up status output |
| `lango onboard` | Guided 5-step setup wizard |
| `lango settings` | Full interactive configuration editor |
| `lango doctor` | Diagnostics and health checks |

Bare `lango` is the default interactive entry point. When the active profile is incomplete, the workbench immediately points the operator to `lango onboard`, `lango settings`, and `lango doctor`. When the profile is ready, the same first screen switches to context-aware starter prompts and exposes the `Enter` / `1-3` quick-start path.

### Dedicated References

Use the dedicated page for command families that have deeper semantics, examples, or operational caveats beyond the quick reference:

| Area | Reference |
|------|-----------|
| Core entrypoints and shared behavior | [core.md](core.md) |
| Status dashboard and dead-letter flows | [status.md](status.md) |
| Agent diagnostics and graph/session inspection | [agent.md](agent.md) |
| Agent memory and graph-adjacent memory flows | [agent-memory.md](agent-memory.md) |
| A2A protocol commands | [a2a.md](a2a.md) |
| Alerts inspection | [alerts.md](alerts.md) |
| Approval inspection | [approval.md](approval.md) |
| Automation (`cron`, `workflow`, `bg`) | [automation.md](automation.md) |
| Configuration management | [config.md](config.md) |
| Contract interaction | [contract.md](contract.md) |
| Economy inspection | [economy.md](economy.md) |
| Extension pack management | [extension.md](extension.md) |
| Graph store management | [graph.md](graph.md) |
| Learning inspection | [learning.md](learning.md) |
| Librarian inspection | [librarian.md](librarian.md) |
| MCP server management | [mcp.md](mcp.md) |
| Metrics and observability | [metrics.md](metrics.md) |
| P2P network management | [p2p.md](p2p.md) |
| Payment operations | [payment.md](payment.md) |
| Provenance and attribution | [provenance.md](provenance.md) |
| RunLedger inspection | [run.md](run.md) |
| Sandbox inspection | [sandbox.md](sandbox.md) |
| Security operations | [security.md](security.md) |
| Smart account management | [smartaccount.md](smartaccount.md) |

### Core Commands

| Command | Description |
|---------|-------------|
| `lango` | Launch standalone mission workbench TUI (default entry point) |
| `lango cockpit` | Launch explicit multi-panel operator dashboard with Mission Control, chat transcript visibility, and operator detail pages |
| `lango chat` | Launch focused chat TUI with tool lifecycle visibility and inline approval controls |
| `lango serve` | Start the gateway server |
| `lango version` | Print version and build info |
| `lango health` | Check gateway health |
| `lango onboard` | Guided 5-step setup wizard |
| `lango settings` | Full interactive configuration editor |
| `lango doctor` | Diagnostics and health checks |

### Status Dashboard

| Command | Description |
|---------|-------------|
| `lango status` | [Unified status dashboard](status.md) (health, config, features) |
| `lango status dead-letter-summary` | Show overview counts, grouped reason/actor/dispatch-family buckets, configurable raw top-N sections, and recent trend windows for the current dead-letter backlog |
| `lango status dead-letters` | List current dead-lettered post-adjudication executions with latest-family and any-match-family filtering |
| `lango status dead-letter <transaction-receipt-id>` | Show detailed dead-letter execution status for one transaction |
| `lango status dead-letter retry <transaction-receipt-id>` | Request retry for one dead-lettered post-adjudication execution with structured follow-up status output |

### Agent Diagnostics

| Command | Description |
|---------|-------------|
| `lango agent trace list` | List recent turn traces with outcomes |
| `lango agent trace show <trace-id>` | Show detailed event timeline for a trace |
| `lango agent graph <session>` | Show delegation graph for a session |
| `lango agent trace metrics` | Per-agent trace-derived performance metrics |

### Config Management

| Command | Description |
|---------|-------------|
| `lango config list` | List all configuration profiles |
| `lango config create <name>` | Create a new profile with defaults or a preset template |
| `lango config use <name>` | Switch to a different profile |
| `lango config delete <name>` | Delete a configuration profile |
| `lango config import <file>` | Import and encrypt a JSON config |
| `lango config export <name>` | Export a profile as plaintext JSON |
| `lango config get <dot.path>` | Read a configuration value by dot-notation path |
| `lango config set <dot.path> [value] [--from-env ENV]` | Set a configuration value by dot-notation path |
| `lango config keys [prefix]` | List available configuration keys |
| `lango config validate` | Validate the active profile |

### Extension Packs

| Command | Description |
|---------|-------------|
| `lango extension inspect <source>` | Print a side-effect-free report about a pack |
| `lango extension install <source>` | Install a pack with inspect + confirm |
| `lango extension list` | List installed extension packs |
| `lango extension remove <name>` | Remove an installed pack |

### Agent & Memory

| Command | Description |
|---------|-------------|
| `lango agent status` | Show agent mode and configuration |
| `lango agent list` | List local and remote agents |
| `lango agent tools` | Show tool category availability from config |
| `lango agent hooks` | Show registered tool hooks |
| `lango memory list` | List observational memory entries |
| `lango memory status` | Show memory system status |
| `lango memory clear` | Clear all memory entries for a session |
| `lango memory agents` | List agents with persistent memory |
| `lango memory agent <name>` | Show memory entries for a specific agent |

`lango agent status` writes through the Cobra command output stream, so wrappers and test harnesses can capture both table and JSON output by replacing `cmd.OutOrStdout()`.
Graph-store commands now live in the dedicated [Graph CLI Reference](graph.md).

### Graph Store

| Command | Description |
|---------|-------------|
| `lango graph status` | Show graph store status |
| `lango graph query` | Query graph triples |
| `lango graph stats` | Show graph statistics |
| `lango graph clear` | Clear all graph data |
| `lango graph add` | Add a triple to the knowledge graph |
| `lango graph export` | Export graph data to a file |
| `lango graph import` | Import graph data from a file |

### Alerts

| Command | Description |
|---------|-------------|
| `lango alerts list` | List recent alerts |
| `lango alerts summary` | Show alert counts by type |

### A2A Protocol

| Command | Description |
|---------|-------------|
| `lango a2a card` | Show local A2A agent card configuration |
| `lango a2a check <url>` | Fetch and display a remote agent card |

### Learning

| Command | Description |
|---------|-------------|
| `lango learning status` | Show learning system configuration |
| `lango learning history` | Show recent learning entries |

### Librarian

| Command | Description |
|---------|-------------|
| `lango librarian status` | Show librarian configuration and inquiry stats |
| `lango librarian inquiries` | List pending knowledge inquiries |

Librarian command output is routed through the Cobra command writer, so wrappers and test harnesses can capture human-readable and JSON output by replacing `cmd.OutOrStdout()`.

### Approval

| Command | Description |
|---------|-------------|
| `lango approval status` | Show approval system configuration |

### Security

| Command | Description |
|---------|-------------|
| `lango security status` | Show security configuration status |
| `lango security change-passphrase` | Rotate the active passphrase without re-encrypting all data |
| `lango security migrate-passphrase` | [DEPRECATED] Legacy full re-encryption passphrase migration |
| `lango security secrets list` | List stored secrets (values hidden) |
| `lango security secrets set <name>` | Store an encrypted secret |
| `lango security secrets delete <name>` | Delete a stored secret |
| `lango security keyring store` | Store passphrase in hardware keyring (Touch ID / TPM) |
| `lango security keyring clear` | Remove passphrase from keyring |
| `lango security keyring status` | Show hardware keyring status |
| `lango security recovery setup` | Set up mnemonic-based passphrase recovery |
| `lango security recovery restore` | Restore access using a recovery mnemonic |
| `lango security db-migrate` | Legacy SQLCipher migration command (unsupported in current runtime) |
| `lango security db-decrypt` | Legacy SQLCipher decrypt command (unsupported in current runtime) |
| `lango security kms status` | Show KMS provider status |
| `lango security kms test` | Test KMS encrypt/decrypt roundtrip |
| `lango security kms keys` | List KMS keys in registry |
| `lango security kms wrap` | Add a KMS KEK slot to protect the master key |
| `lango security kms detach` | Remove a KMS KEK slot from the envelope |

### Payment

| Command | Description |
|---------|-------------|
| `lango payment balance` | Show USDC wallet balance |
| `lango payment history` | Show payment transaction history |
| `lango payment limits` | Show spending limits and daily usage |
| `lango payment info` | Show wallet and payment system info |
| `lango payment send` | Send a USDC payment |
| `lango payment x402` | Show X402 auto-pay configuration |

### P2P Network

| Command | Description |
|---------|-------------|
| `lango p2p status` | Show P2P node status |
| `lango p2p peers` | List connected peers |
| `lango p2p connect <multiaddr>` | Connect to a peer by multiaddr |
| `lango p2p disconnect <peer-id>` | Disconnect from a peer |
| `lango p2p firewall list` | List firewall ACL rules |
| `lango p2p firewall add` | Add a firewall ACL rule |
| `lango p2p firewall remove` | Remove firewall rules for a peer |
| `lango p2p discover` | Discover agents by capability |
| `lango p2p identity` | Show local DID and peer identity |
| `lango p2p reputation --peer-did <did>` | Query peer trust score |
| `lango p2p pricing` | Show tool pricing |
| `lango p2p workspace create <name>` | Create a local collaborative workspace |
| `lango p2p workspace list` | List local collaborative workspaces |
| `lango p2p workspace status <workspace-id>` | Show one local collaborative workspace |
| `lango p2p workspace join <workspace-id>` | Join a local collaborative workspace |
| `lango p2p workspace leave <workspace-id>` | Leave a local collaborative workspace |
| `lango p2p git init <workspace-id>` | Describe how to initialize a workspace git repository |
| `lango p2p git log <workspace-id>` | Describe how to inspect workspace commit history |
| `lango p2p git diff <workspace-id> <from> <to>` | Describe how to diff workspace commits |
| `lango p2p git push <workspace-id>` | Describe how to push a workspace git bundle to peers |
| `lango p2p git fetch <workspace-id>` | Describe how to fetch a workspace git bundle from peers |
| `lango p2p provenance push <peer-did> <session-key>` | Push a signed provenance bundle to a peer |
| `lango p2p provenance fetch <peer-did> <session-key>` | Fetch and import a signed provenance bundle from a peer |
| `lango p2p session list` | List active peer sessions |
| `lango p2p session revoke` | Revoke a peer session |
| `lango p2p session revoke-all` | Revoke all active peer sessions |
| `lango p2p sandbox status` | Show sandbox runtime status |
| `lango p2p sandbox test` | Run sandbox smoke test |
| `lango p2p sandbox cleanup` | Remove orphaned sandbox containers |
| `lango p2p team list` | Describe how to inspect active P2P teams |
| `lango p2p team status <id>` | Describe how to inspect runtime-backed team status |
| `lango p2p team disband <id>` | Describe how to disband a runtime-backed team |
| `lango p2p zkp status` | Show ZKP configuration |
| `lango p2p zkp circuits` | List compiled ZKP circuits |

### Economy

| Command | Description |
|---------|-------------|
| `lango economy budget status` | Show budget allocation status |
| `lango economy risk status` | Show risk assessment configuration |
| `lango economy pricing status` | Show dynamic pricing configuration |
| `lango economy negotiate status` | Show negotiation protocol status |
| `lango economy escrow status` | Show escrow service status |
| `lango economy escrow list` | Show escrow configuration summary |
| `lango economy escrow show` | Show detailed escrow configuration |
| `lango economy escrow sentinel status` | Show escrow sentinel status |

### Smart Account

| Command | Description |
|---------|-------------|
| `lango account info` | Show smart account configuration and status |
| `lango account deploy` | Deploy a new Safe smart account with ERC-7579 adapter |
| `lango account session list` | List active session keys |
| `lango account session create` | Create a new session key |
| `lango account session revoke` | Revoke a session key or all session keys |
| `lango account module list` | List registered ERC-7579 modules |
| `lango account module install` | Install an ERC-7579 module |
| `lango account policy show` | Show current harness policy configuration |
| `lango account policy set` | Set harness policy limits |
| `lango account paymaster status` | Show paymaster configuration and approval status |
| `lango account paymaster approve` | Approve USDC spending for the paymaster |

### Contract

| Command | Description |
|---------|-------------|
| `lango contract read` | Call a view/pure smart contract method |
| `lango contract call` | Execute a state-changing contract method |
| `lango contract abi load` | Load and cache a contract ABI |

### Metrics

| Command | Description |
|---------|-------------|
| `lango metrics` | Show system metrics snapshot |
| `lango metrics sessions` | Show per-session token usage |
| `lango metrics tools` | Show per-tool metrics |
| `lango metrics agents` | Show per-agent metrics |
| `lango metrics policy` | Show policy decision statistics |
| `lango metrics history` | Show historical metrics |

### Automation

| Command | Description |
|---------|-------------|
| `lango cron add` | Add a new cron job |
| `lango cron list` | List all cron jobs |
| `lango cron delete <id-or-name>` | Delete a cron job |
| `lango cron pause <id-or-name>` | Pause a cron job |
| `lango cron resume <id-or-name>` | Resume a paused cron job |
| `lango cron history` | Show cron execution history |
| `lango workflow run <file>` | Execute a workflow YAML file |
| `lango workflow list` | List workflow runs |
| `lango workflow status <run-id>` | Show workflow run status |
| `lango workflow cancel <run-id>` | Cancel a running workflow |
| `lango workflow history` | Show workflow execution history |
| `lango workflow validate <file>` | Validate a workflow YAML file |
| `lango bg list` | List background tasks |
| `lango bg status <id>` | Show background task status |
| `lango bg cancel <id>` | Cancel a running background task |
| `lango bg result <id>` | Show completed task result |

### MCP Servers

| Command | Description |
|---------|-------------|
| `lango mcp list` | List all configured MCP servers |
| `lango mcp add <name>` | Add a new MCP server |
| `lango mcp remove <name>` | Remove an MCP server configuration |
| `lango mcp get <name>` | Show server details and discovered tools |
| `lango mcp test <name>` | Test server connectivity |
| `lango mcp enable <name>` | Enable an MCP server |
| `lango mcp disable <name>` | Disable an MCP server |

### RunLedger (Task OS)

!!! warning "Experimental"
    The RunLedger is experimental. See [RunLedger](../features/run-ledger.md).

| Command | Description |
|---------|-------------|
| `lango run list` | List recent runs |
| `lango run status` | Show RunLedger configuration |
| `lango run journal <run-id>` | View run journal events |

### Session Provenance

!!! warning "Experimental"
    Session provenance is experimental. See [Session Provenance](../features/provenance.md).

| Command | Description |
|---------|-------------|
| `lango provenance status` | Show provenance configuration and state |
| `lango provenance checkpoint list --run <id>` | List checkpoints |
| `lango provenance checkpoint create <label> --run <id>` | Create a manual checkpoint |
| `lango provenance checkpoint show <id>` | Show checkpoint details |
| `lango provenance session tree <session-key>` | Show session hierarchy tree |
| `lango provenance session list` | List persisted session nodes |
| `lango provenance attribution show <session-key>` | Show attribution data for a session |
| `lango provenance attribution report <session-key>` | Generate attribution report |
| `lango provenance bundle export <session-key>` | Export a signed provenance bundle |
| `lango provenance bundle import <file>` | Import a signed provenance bundle |

### Sandbox (OS-level)

!!! warning "Experimental"
    The OS-level sandbox is experimental. This is distinct from `lango p2p sandbox` which manages P2P remote execution isolation.

| Command | Description |
|---------|-------------|
| `lango sandbox status` | Show sandbox configuration and platform capabilities |
| `lango sandbox test` | Run OS sandbox smoke tests |

## Global Behavior

All commands read configuration from the active encrypted profile stored in `~/.lango/lango.db`. On first run, Lango prompts for a passphrase to initialize encryption.

Commands that need a running server (like `lango health`) connect to `localhost` on the configured port (default: `18789`).

!!! tip "Getting Started"
    If you're new to Lango, start with `lango onboard` to walk through the initial setup, then use `lango doctor` to verify everything is configured correctly.
