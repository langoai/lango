## Context

The P2P networking core is fully implemented across `internal/p2p/` (node, identity, handshake, firewall, discovery, protocol, ZKP) but has zero user-facing surface. Users cannot inspect, configure, or interact with P2P features from the CLI, and agents have no awareness of P2P tools in their prompts. Documentation does not mention P2P capabilities.

This change adds the presentation layer: CLI commands, agent prompts, embedded skills, and documentation.

## Goals / Non-Goals

**Goals:**
- Expose all P2P core functionality through `lango p2p` CLI commands
- Make the agent aware of P2P tools via updated prompts
- Provide embedded skills for common P2P operations
- Document P2P features, CLI commands, and configuration

**Non-Goals:**
- Modifying P2P core behavior (node, handshake, firewall logic)
- Adding new P2P protocol features
- Creating agent tools (tool registration in `internal/tools/p2p/`) — that is a separate change
- TUI integration for P2P settings

## Decisions

### D1: CLI follows bootstrap Result loader pattern
**Decision**: Use `bootLoader func() (*bootstrap.Result, error)` pattern from `cli/payment/`.
**Rationale**: P2P Node requires config from bootstrap. Consistent with existing CLI patterns. The `initP2PDeps` function creates a temporary P2P node for the duration of the CLI command.
**Alternative considered**: Config-only loader (like `cli/memory/`) — rejected because P2P commands need a live libp2p host to query peers and connect.

### D2: CLI creates its own P2P Node instance
**Decision**: Each CLI invocation creates and starts its own P2P node via `p2p.NewNode()`, then stops it on cleanup.
**Rationale**: CLI commands are short-lived and independent of `lango serve`. Creating a dedicated node avoids IPC complexity with a running server.
**Trade-off**: The CLI node won't see peers connected to the server's node. Commands like `peers` show the CLI node's connections, not the server's. This is acceptable for the initial implementation.

### D3: Firewall CLI reads config, not runtime state
**Decision**: `lango p2p firewall list` reads rules from `P2PConfig.FirewallRules`, not from a running firewall instance.
**Rationale**: The CLI-created node has a fresh firewall with only config rules. Runtime-added rules (via agent tools) only exist in the server process. Config-based listing is the correct behavior for a CLI inspection tool.

### D4: Prompt changes are additive only
**Decision**: Add a 10th tool category and new TOOL_USAGE section without restructuring existing content.
**Rationale**: Minimizes risk of breaking existing prompt behavior. The P2P section follows the same format as existing tool sections.

### D5: Skills use script type with direct CLI mapping
**Decision**: Each skill is a simple `type: script` that maps to `lango p2p <subcommand>`.
**Rationale**: Matches the pattern of all 30 existing embedded skills. No composite or template skills needed since each P2P operation maps to a single CLI command.

## Risks / Trade-offs

- **[CLI node isolation]** CLI P2P commands operate on a fresh node, not the server's node → Users may be confused when `lango p2p peers` shows different results than what the running server sees. → Mitigation: Document this behavior clearly; future work can add IPC to query the server.
- **[Firewall add is runtime-only]** `lango p2p firewall add` prints a message but cannot persist rules → Mitigation: Output includes guidance to edit config for persistence.
- **[Discovery requires bootstrap peers]** `lango p2p discover` on a fresh node finds no peers without pre-configured bootstrap peers → Mitigation: Command output includes guidance when no agents are found.
