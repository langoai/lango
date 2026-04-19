# P2P Commands

Commands for managing the P2P network on the Sovereign Agent Network. P2P must be enabled in configuration (`p2p.enabled = true`). See the [P2P Network](../features/p2p-network.md) section for detailed documentation.

```
lango p2p <subcommand>
```

!!! warning "Experimental Feature"
    The P2P networking system is experimental. Protocol and behavior may change between releases.

---

## lango p2p status

Show the P2P node status including peer ID, listen addresses, connected peer count, and feature flags.

```
lango p2p status [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango p2p status
P2P Node Status
  Peer ID:          QmYourPeerId123...
  Listen Addrs:     [/ip4/0.0.0.0/tcp/9000]
  Connected Peers:  3 / 50
  mDNS:             true
  Relay:            false
  ZK Handshake:     false
```

---

## lango p2p peers

List all currently connected peers with their peer IDs and multiaddrs.

```
lango p2p peers [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango p2p peers
PEER ID                          ADDRESS
QmPeer1abc123...                 /ip4/192.168.1.5/tcp/9000
QmPeer2def456...                 /ip4/10.0.0.3/tcp/9001
```

---

## lango p2p connect

Connect to a peer by its full multiaddr (including the `/p2p/<peer-id>` suffix).

```
lango p2p connect <multiaddr>
```

| Argument | Description |
|----------|-------------|
| `multiaddr` | Full multiaddr of the peer (e.g., `/ip4/1.2.3.4/tcp/9000/p2p/QmPeerId`) |

**Example:**

```bash
$ lango p2p connect /ip4/192.168.1.5/tcp/9000/p2p/QmPeer1abc123
Connected to peer QmPeer1abc123
```

---

## lango p2p disconnect

Disconnect from a peer by its peer ID.

```
lango p2p disconnect <peer-id>
```

| Argument | Description |
|----------|-------------|
| `peer-id` | Peer ID to disconnect from |

**Example:**

```bash
$ lango p2p disconnect QmPeer1abc123
Disconnected from peer QmPeer1abc123
```

---

## lango p2p firewall

Manage knowledge firewall ACL rules that control peer access.

### lango p2p firewall list

List all configured firewall ACL rules.

```
lango p2p firewall list [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango p2p firewall list
PEER DID                              ACTION  TOOLS         RATE LIMIT
did:lango:02abc...                    allow   search_*      10/min
*                                     deny    exec_*        unlimited
```

### lango p2p firewall add

Add a new firewall ACL rule (runtime only — persist by updating configuration).

```
lango p2p firewall add --peer-did <did> --action <allow|deny> [--tools <patterns>] [--rate-limit <n>]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--peer-did` | string | *required* | Peer DID to apply the rule to (`"*"` for all) |
| `--action` | string | `allow` | Rule action: `allow` or `deny` |
| `--tools` | []string | `[]` | Tool name patterns (empty = all tools) |
| `--rate-limit` | int | `0` | Max requests per minute (0 = unlimited) |

**Example:**

```bash
$ lango p2p firewall add --peer-did "did:lango:02abc..." --action allow --tools "search_*,rag_*" --rate-limit 10
Firewall rule added (runtime only):
  Peer DID:    did:lango:02abc...
  Action:      allow
  Tools:       search_*, rag_*
  Rate Limit:  10/min
```

### lango p2p firewall remove

Remove all firewall rules matching a peer DID.

```
lango p2p firewall remove <peer-did>
```

| Argument | Description |
|----------|-------------|
| `peer-did` | Peer DID whose rules should be removed |

---

## lango p2p discover

Discover agents on the P2P network via GossipSub. Optionally filter by capability tag.

```
lango p2p discover [--tag <capability>] [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tag` | string | `""` | Filter by capability tag |
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango p2p discover --tag research
NAME              DID                     CAPABILITIES          PEER ID
research-bot      did:lango:02abc...      research, summarize   QmPeer1abc123
```

---

## lango p2p identity

Show the local P2P identity including the active DID when available, peer ID, key storage mode, and listen addresses.

Lango supports both legacy wallet-derived `did:lango:<hex>` identities and bundle-backed `did:lango:v2:<hash>` identities. The CLI and `GET /api/p2p/identity` expose the active DID when available. The `/api/p2p/*` routes are public only when gateway auth is disabled; otherwise the subtree is protected by gateway auth.

```
lango p2p identity [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango p2p identity
P2P Identity
  DID:          did:lango:v2:abcdef1234567890abcdef1234567890abcdef12
  Peer ID:      QmYourPeerId123...
  Key Storage:  secrets-store
  Listen Addrs:
    /ip4/0.0.0.0/tcp/9000
    /ip6/::/tcp/9000
```

---

## `lango p2p reputation`

Show peer reputation and trust score details.

### Usage

```bash
lango p2p reputation --peer-did <did> [--json]
```

### Flags

| Flag | Description |
|------|-------------|
| `--peer-did` | The DID of the peer to query (required) |
| `--json` | Output as JSON |

### Examples

```bash
# Show reputation for a peer
lango p2p reputation --peer-did "did:lango:abc123"

# Output as JSON
lango p2p reputation --peer-did "did:lango:abc123" --json
```

### Output Fields

| Field | Description |
|-------|-------------|
| Trust Score | Current trust score (0.0 to 1.0) |
| Successes | Number of successful exchanges |
| Failures | Number of failed exchanges |
| Timeouts | Number of timed-out exchanges |
| First Seen | Timestamp of first interaction |
| Last Interaction | Timestamp of most recent interaction |

---

## lango p2p session

Manage P2P sessions. List, revoke, or revoke all authenticated peer sessions.

### lango p2p session list

List all active (non-expired, non-invalidated) peer sessions.

```
lango p2p session list [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango p2p session list
PEER DID                              CREATED                    EXPIRES                    ZK VERIFIED
did:lango:02abc123...                 2026-02-25T10:00:00Z       2026-02-25T11:00:00Z       true
did:lango:03def456...                 2026-02-25T10:30:00Z       2026-02-25T11:30:00Z       false
```

---

### lango p2p session revoke

Revoke a specific peer's session by DID.

```
lango p2p session revoke --peer-did <did>
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--peer-did` | string | *required* | The DID of the peer to revoke |

**Example:**

```bash
$ lango p2p session revoke --peer-did "did:lango:02abc123..."
Session for did:lango:02abc123... revoked.
```

---

### lango p2p session revoke-all

Revoke all active peer sessions.

```
lango p2p session revoke-all
```

**Example:**

```bash
$ lango p2p session revoke-all
All sessions revoked.
```

---

## lango p2p provenance

Exchange signed provenance bundles with peers through the running gateway.

These commands are server-backed. They require:

- `lango serve` to be running
- `p2p.enabled = true`
- an active authenticated session for the target peer DID

### lango p2p provenance push

```bash
lango p2p provenance push <peer-did> <session-key> [--redaction <none|content|full>] [--addr <gateway>]
```

Push a signed provenance bundle for the given session key to a remote peer.

### lango p2p provenance fetch

```bash
lango p2p provenance fetch <peer-did> <session-key> [--redaction <none|content|full>] [--addr <gateway>]
```

Fetch a signed provenance bundle from a remote peer and verify-and-store import it locally.

---

## lango p2p sandbox

Manage the P2P tool execution sandbox. Inspect sandbox status, run smoke tests, and clean up orphaned containers.

### lango p2p sandbox status

Show the current sandbox runtime status including isolation configuration, container mode, and active runtime.

```
lango p2p sandbox status
```

**Example (subprocess mode):**

```bash
$ lango p2p sandbox status
Tool isolation: enabled
  Timeout per tool: 30s
  Max memory (MB):  512
  Container mode:   disabled (subprocess fallback)
```

**Example (container mode):**

```bash
$ lango p2p sandbox status
Tool isolation: enabled
  Timeout per tool: 30s
  Max memory (MB):  512
  Container mode:   enabled
  Runtime config:   auto
  Image:            lango-sandbox:latest
  Network mode:     none
  Active runtime:   docker
  Pool size:        3
```

---

### lango p2p sandbox test

Run a sandbox smoke test by executing a simple echo tool through the sandbox.

```
lango p2p sandbox test
```

**Example:**

```bash
$ lango p2p sandbox test
Using container runtime: docker
Smoke test passed: map[msg:sandbox-smoke-test]
```

---

### lango p2p sandbox cleanup

Find and remove orphaned Docker containers with the `lango.sandbox=true` label.

```
lango p2p sandbox cleanup
```

**Example:**

```bash
$ lango p2p sandbox cleanup
Orphaned sandbox containers cleaned up.
```

---

## `lango p2p pricing`

Show provider-side P2P quote configuration.

### Usage

```bash
lango p2p pricing [--tool <name>] [--json]
```

### Flags

| Flag | Description |
|------|-------------|
| `--tool` | Filter pricing for a specific tool |
| `--json` | Output as JSON |

### Examples

```bash
# Show all pricing
lango p2p pricing

# Show pricing for a specific tool
lango p2p pricing --tool "knowledge_search"

# Output as JSON
lango p2p pricing --json
```

---

## lango p2p team

Inspect the current team operator surface for the running P2P runtime. Teams are real runtime-only coordination structures, but the CLI commands below currently act as truth-aligned guidance rather than full live team control. See the [P2P Teams](../features/p2p-network.md#p2p-team-coordination) section for subsystem details.

### lango p2p team list

Inspect what the CLI currently reports for active P2P teams.

```
lango p2p team list [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango p2p team list
No active teams.

Teams are runtime-only structures created during agent collaboration.
Start the server with 'lango serve' and inspect/form teams via runtime integrations and agent tools.
```

### lango p2p team status

Inspect how the CLI currently guides operators toward live team inspection.

```
lango p2p team status <team-id> [--json]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `team-id` | Yes | Team ID |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango p2p team status a1b2c3d4-5678-9012-abcd-ef1234567890
Team not found.

Teams are runtime-only structures that exist only while the server is running.
Use the running server plus the team runtime or agent tools for live inspection.
```

### lango p2p team disband

Inspect how the CLI currently guides operators toward live team disband.

```
lango p2p team disband <team-id>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `team-id` | Yes | Team ID to disband |

**Example:**

```bash
$ lango p2p team disband a1b2c3d4-5678-9012-abcd-ef1234567890
Team not found.

Teams are runtime-only structures.
Use the running server plus the team runtime or agent tools to disband a live team.
```

### Team Coordination Features

Teams support configurable conflict resolution and payment coordination:

- **Conflict Resolution**: `trust_weighted` (default), `majority_vote`, `leader_decides`, `fail_on_conflict`
- **Assignment**: `best_match`, `round_robin`, `load_balanced`
- **Payment Modes**: Trust-based mode selection — `free` (price=0), `postpay` (trust >= 0.8), `prepay` (trust < 0.8)

Teams are runtime-only structures managed by the running server. Today the stable operator path is runtime-backed or tool-backed (`team_form`, `team_delegate`, `team_status`, `team_list`, `team_disband`), while these CLI commands remain guidance-oriented.

See the [P2P Team Coordination](../features/p2p-network.md#p2p-team-coordination) section for detailed documentation on conflict resolution strategies, assignment strategies, and payment coordination.

---

## lango p2p zkp

Inspect ZKP (zero-knowledge proof) configuration and compiled circuits. See the [ZKP](../features/zkp.md) section for details.

### lango p2p zkp status

Show ZKP configuration, including proving scheme, SRS mode, and compiled circuit count.

```
lango p2p zkp status [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango p2p zkp status
ZKP Status
  Proving Scheme:   plonk
  SRS Mode:         unsafe
  ZK Handshake:     true
  ZK Attestation:   true
  Compiled Circuits: 4
```

### lango p2p zkp circuits

List all available ZKP circuits and their compilation status.

```
lango p2p zkp circuits [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango p2p zkp circuits
CIRCUIT              COMPILED  CONSTRAINTS  SCHEME
ownership            true      245          plonk
capability           true      512          plonk
balance_range        true      128          plonk
attestation          true      389          plonk
```

---

## lango p2p workspace

Inspect the current workspace operator surface for the running P2P runtime. Workspaces are real runtime structures, but the CLI commands below mostly point operators toward server-backed or tool-backed flows rather than performing full live control directly. See the [Collaborative Workspaces](../features/p2p-network.md#collaborative-workspaces) section for details.

### lango p2p workspace create

Inspect how the CLI currently guides operators toward live workspace creation.

```
lango p2p workspace create <name> [--goal <goal>] [--json]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Workspace name |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--goal` | string | `""` | Description of the workspace goal |
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango p2p workspace create "research-project" --goal "Collaborative research on RAG optimization"
Workspace creation requires a running server.
Start the server with 'lango serve' and use the server-backed runtime or agent tools.

Example: p2p_workspace_create name="research-project" goal="Collaborative research on RAG optimization"
```

### lango p2p workspace list

Inspect what the CLI currently reports for runtime-backed workspaces.

```
lango p2p workspace list [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango p2p workspace list
No workspaces found.

Workspaces are runtime structures managed by the running server.
Start the server with 'lango serve' and use the server-backed runtime or p2p_workspace_* tools.
```

### lango p2p workspace status

Inspect how the CLI currently guides operators toward live workspace inspection.

```
lango p2p workspace status <id> [--json]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `id` | Yes | Workspace ID |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango p2p workspace status a1b2c3d4-5678-9012-abcd-ef1234567890
Workspace not found.

Workspaces are runtime structures.
Use the running server plus workspace runtime integrations or agent tools for inspection.
```

### lango p2p workspace join

Inspect how the CLI currently guides operators toward live workspace join.

```
lango p2p workspace join <id>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `id` | Yes | Workspace ID to join |

**Example:**

```bash
$ lango p2p workspace join a1b2c3d4-5678-9012-abcd-ef1234567890
Joining a workspace requires a running server.
Use 'lango serve' and the server-backed runtime or p2p_workspace_join tool.
```

### lango p2p workspace leave

Inspect how the CLI currently guides operators toward live workspace leave.

```
lango p2p workspace leave <id>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `id` | Yes | Workspace ID to leave |

**Example:**

```bash
$ lango p2p workspace leave a1b2c3d4-5678-9012-abcd-ef1234567890
Leaving a workspace requires a running server.
Use 'lango serve' and the server-backed runtime or p2p_workspace_leave tool.
```

### Workspace Features

Workspaces support configurable collaboration features:

- **Lifecycle**: `forming` → `active` → `archived`
- **Message Types**: `TASK_PROPOSAL`, `LOG_STREAM`, `COMMIT_SIGNAL`, `KNOWLEDGE_SHARE`, `MEMBER_JOINED`, `MEMBER_LEFT`
- **Chronicler**: Chronicler hooks exist, but graph-triple persistence depends on the triple-adder wiring and is not yet the default live operator path
- **Contribution Tracking**: Per-agent metrics (commits, code bytes, messages)
- **Auto Sandbox**: Optionally isolate workspace operations in sandboxed environments

Workspaces are runtime structures managed by the running server. Today the stable operator path is server-backed or tool-backed (`p2p_workspace_create`, `p2p_workspace_join`, `p2p_workspace_leave`), while these CLI commands remain guidance-oriented.

---

## lango p2p git

Inspect the current git bundle operator surface for workspace code collaboration. The git bundle runtime is real, but these CLI commands mostly point operators toward server-backed or tool-backed flows instead of providing full direct live repository control.

### lango p2p git init

Inspect how the CLI currently guides operators toward live workspace git initialization.

```
lango p2p git init <workspace-id>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `workspace-id` | Yes | Workspace ID |

**Example:**

```bash
$ lango p2p git init a1b2c3d4-5678-9012-abcd-ef1234567890
Git init requires a running server.
Use 'lango serve' and the runtime API or p2p_git_init tool.
```

### lango p2p git log

Inspect how the CLI currently guides operators toward live workspace git history.

```
lango p2p git log <workspace-id> [--limit <n>] [--json]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `workspace-id` | Yes | Workspace ID |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit` | int | `20` | Maximum number of commits to show |
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango p2p git log a1b2c3d4-5678-9012-abcd-ef1234567890 --limit 5
No commits found.
Git operations require a running server with workspace enabled.
Use the runtime API or p2p_git_* tools for live repository inspection.
```

### lango p2p git diff

Inspect how the CLI currently guides operators toward live workspace diffs.

```
lango p2p git diff <workspace-id> <from> <to>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `workspace-id` | Yes | Workspace ID |
| `from` | Yes | Source commit hash |
| `to` | Yes | Target commit hash |

**Example:**

```bash
$ lango p2p git diff a1b2c3d4-... abc1234 def5678
Diff requires a running server.
Use 'lango serve' and the runtime API or p2p_git_diff tool.
```

### lango p2p git push

Inspect how the CLI currently guides operators toward live git bundle push.

```
lango p2p git push <workspace-id>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `workspace-id` | Yes | Workspace ID |

**Example:**

```bash
$ lango p2p git push a1b2c3d4-5678-9012-abcd-ef1234567890
Push requires a running server.
Use 'lango serve' and the server-backed runtime or p2p_git_push tool.
```

### lango p2p git fetch

Inspect how the CLI currently guides operators toward live git bundle fetch.

```
lango p2p git fetch <workspace-id>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `workspace-id` | Yes | Workspace ID |

**Example:**

```bash
$ lango p2p git fetch a1b2c3d4-5678-9012-abcd-ef1234567890
Fetch requires a running server.
Use 'lango serve' and the server-backed runtime plus provenance or workspace artifact tools for live exchange.
```

### Incremental Bundles

Instead of transferring the entire repository history every time, Lango supports incremental bundles that contain only commits after a known base commit. This significantly reduces transfer size for active workspaces.

The `CreateIncrementalBundle` operation takes a base commit hash and produces a bundle containing only `baseCommit..HEAD`. If the base commit is not found in the repository, it falls back to a full bundle automatically.

Before applying a received bundle, `VerifyBundle` checks that the bundle's prerequisite commits exist in the local repo. `SafeApplyBundle` combines verification, ref snapshot, application, and automatic rollback on failure into a single atomic operation:

1. **Verify** — check prerequisites are present
2. **Snapshot** — capture current ref state
3. **Apply** — unbundle into the repository
4. **Rollback** — if apply fails, restore refs from the snapshot

`HasCommit` checks whether a specific commit exists locally, which is useful for determining the correct base commit before requesting an incremental bundle from a peer.

---

### Workflow Example: Git Bundle Exchange

A typical runtime-backed collaboration workflow using git bundles:

```bash
# 1. Start the runtime that owns the live workspace repo
lango serve

# 2. Use the workspace/git tools to initialize and exchange bundles
p2p_git_init workspace_id="a1b2c3d4-..."
p2p_git_push workspace_id="a1b2c3d4-..."
p2p_git_fetch workspace_id="a1b2c3d4-..."

# 3. Use the guidance commands when you need CLI reminders about the operator path
lango p2p git log a1b2c3d4-... --limit 10
lango p2p git diff a1b2c3d4-... abc1234 def5678
```

### Git Bundle Features

- **Bare Repositories**: Each workspace has an isolated bare git repo at `~/.lango/workspaces/<id>/repo.git`
- **Bundle Protocol**: Uses `git bundle create/unbundle` for atomic transfers over the P2P network
- **Incremental Bundles**: Transfer only new commits since a known base, with automatic full-bundle fallback
- **Safe Apply**: Verify prerequisites, snapshot refs, apply, and auto-rollback on failure
- **Task Branches** (library): Per-task isolation via `task/{taskID}` branches, managed by agent tools at runtime
- **Conflict Detection** (library): `git merge-tree --write-tree` detects merge conflicts without a working tree
- **DAG Leaf Detection**: Identifies leaf commits (no children) for conflict detection
- **Size Limits**: Configurable `maxBundleSizeBytes` to prevent oversized transfers

Git operations require the `git` binary to be installed and available in PATH. In this command family, live provenance exchange remains the main concrete CLI-backed exception; workspace and git bundle control are otherwise primarily server-backed or tool-backed today.
