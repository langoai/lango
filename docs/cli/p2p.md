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

Show the local P2P identity including peer ID, key directory, and listen addresses.

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
  Peer ID:      QmYourPeerId123...
  Key Dir:      ~/.lango/p2p
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

Show P2P tool pricing configuration.

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

Manage P2P teams — task-scoped collaboration groups between agents across the network. See the [P2P Teams](../features/p2p-network.md#p2p-team-coordination) section for details.

### lango p2p team list

List all active P2P teams.

```
lango p2p team list [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango p2p team list
ID                                    STATUS    MEMBERS  LEADER DID                   TASK
a1b2c3d4-5678-9012-abcd-ef1234567890  active    3        did:lango:02abc...          Research project
e5f6g7h8-9012-3456-cdef-ab1234567890  forming   2        did:lango:03def...          Code review
```

### lango p2p team status

Show detailed status for a specific team, including members and their roles.

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
Team: a1b2c3d4-5678-9012-abcd-ef1234567890
  Status:  active
  Task:    Research project

Members:
  DID                          ROLE       STATUS
  did:lango:02abc...           leader     idle
  did:lango:03def...           worker     busy
  did:lango:04ghi...           reviewer   idle
```

### lango p2p team disband

Disband an active team. Only the team leader can disband.

```
lango p2p team disband <team-id>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `team-id` | Yes | Team ID to disband |

**Example:**

```bash
$ lango p2p team disband a1b2c3d4-5678-9012-abcd-ef1234567890
Team a1b2c3d4-5678-9012-abcd-ef1234567890 disbanded.
```

### Team Coordination Features

Teams support configurable conflict resolution and payment coordination:

- **Conflict Resolution**: `trust_weighted` (default), `majority_vote`, `leader_decides`, `fail_on_conflict`
- **Assignment**: `best_match`, `round_robin`, `load_balanced`
- **Payment Modes**: Trust-based mode selection — `free` (price=0), `postpay` (trust >= 0.7), `prepay` (trust < 0.7)

Teams are runtime-only structures managed by the running server. Use `lango serve` to start the server and form teams via the agent tools (`p2p_team_create`, `p2p_team_join`).

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

Manage P2P workspaces — collaborative environments where agents share code, messages, and git bundles. See the [Collaborative Workspaces](../features/p2p-network.md#collaborative-workspaces) section for details.

### lango p2p workspace create

Create a new collaborative workspace.

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
Workspace created: a1b2c3d4-5678-9012-abcd-ef1234567890
  Name:   research-project
  Status: forming
  Goal:   Collaborative research on RAG optimization
```

### lango p2p workspace list

List all known workspaces.

```
lango p2p workspace list [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango p2p workspace list
ID                                    STATUS    MEMBERS  NAME
a1b2c3d4-5678-9012-abcd-ef1234567890  active    3        research-project
e5f6g7h8-9012-3456-cdef-ab1234567890  forming   1        code-review
```

### lango p2p workspace status

Show detailed workspace status including members and contributions.

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
Workspace: a1b2c3d4-5678-9012-abcd-ef1234567890
  Name:    research-project
  Status:  active
  Goal:    Collaborative research on RAG optimization

Members:
  DID                          ROLE       JOINED
  did:lango:02abc...           creator    2026-03-10T09:00:00Z
  did:lango:03def...           member     2026-03-10T09:15:00Z
```

### lango p2p workspace join

Join an existing workspace.

```
lango p2p workspace join <id>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `id` | Yes | Workspace ID to join |

**Example:**

```bash
$ lango p2p workspace join a1b2c3d4-5678-9012-abcd-ef1234567890
Joined workspace a1b2c3d4-5678-9012-abcd-ef1234567890.
```

### lango p2p workspace leave

Leave a workspace.

```
lango p2p workspace leave <id>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `id` | Yes | Workspace ID to leave |

**Example:**

```bash
$ lango p2p workspace leave a1b2c3d4-5678-9012-abcd-ef1234567890
Left workspace a1b2c3d4-5678-9012-abcd-ef1234567890.
```

### Workspace Features

Workspaces support configurable collaboration features:

- **Lifecycle**: `forming` → `active` → `archived`
- **Message Types**: `TASK_PROPOSAL`, `LOG_STREAM`, `COMMIT_SIGNAL`, `KNOWLEDGE_SHARE`, `MEMBER_JOINED`, `MEMBER_LEFT`
- **Chronicler**: Persists workspace messages as graph triples for knowledge retention
- **Contribution Tracking**: Per-agent metrics (commits, code bytes, messages)
- **Auto Sandbox**: Optionally isolate workspace operations in sandboxed environments

Workspaces are runtime structures managed by the running server. Use `lango serve` to start the server and create workspaces via the agent tools (`p2p_workspace_create`, `p2p_workspace_join`).

---

## lango p2p git

Manage git bundle exchange for workspace code collaboration. Workspaces use bare git repositories with bundle-based transfer for atomic code sharing.

### lango p2p git init

Initialize a bare git repository for a workspace.

```
lango p2p git init <workspace-id>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `workspace-id` | Yes | Workspace ID |

**Example:**

```bash
$ lango p2p git init a1b2c3d4-5678-9012-abcd-ef1234567890
Initialized bare repo for workspace a1b2c3d4-5678-9012-abcd-ef1234567890.
```

### lango p2p git log

Show recent commits in a workspace repository.

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
HASH       AUTHOR          TIMESTAMP                MESSAGE
abc1234    agent-alpha     2026-03-10T10:30:00Z     Add RAG optimization module
def5678    agent-beta      2026-03-10T10:15:00Z     Initial project structure
```

### lango p2p git diff

Show diff between two commits in a workspace repository.

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
diff --git a/rag/optimizer.go b/rag/optimizer.go
...
```

### lango p2p git push

Create a git bundle from the workspace repository and push to peers.

```
lango p2p git push <workspace-id>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `workspace-id` | Yes | Workspace ID |

**Example:**

```bash
$ lango p2p git push a1b2c3d4-5678-9012-abcd-ef1234567890
Bundle created (12.4 KB), HEAD: abc1234
Pushed to 2 workspace peers.
```

### lango p2p git fetch

Fetch a git bundle from workspace peers and apply to the local repository.

```
lango p2p git fetch <workspace-id>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `workspace-id` | Yes | Workspace ID |

**Example:**

```bash
$ lango p2p git fetch a1b2c3d4-5678-9012-abcd-ef1234567890
Fetched bundle from did:lango:03def... (8.2 KB)
Applied 3 new commits.
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

### lango p2p git branch

Manage task branches within a workspace repository. Task branches use the `task/{taskID}` naming convention and support the full lifecycle: create, list, merge, and delete.

#### lango p2p git branch create

Create a task branch in the workspace repository.

```
lango p2p git branch create <workspace-id> --task <task-id> [--base <branch>]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--task` | string | *required* | Task ID (branch will be named `task/{taskID}`) |
| `--base` | string | `HEAD` | Base branch to create from |

The operation is idempotent — if the branch already exists, it succeeds without error.

**Example:**

```bash
$ lango p2p git branch create a1b2c3d4-... --task TASK-42
Created branch task/TASK-42 in workspace a1b2c3d4-...
```

#### lango p2p git branch list

List all branches in the workspace repository.

```
lango p2p git branch list <workspace-id> [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango p2p git branch list a1b2c3d4-...
NAME              COMMIT    HEAD  UPDATED
main              abc1234   *     2026-03-10T10:30:00Z
task/TASK-42      def5678         2026-03-10T11:00:00Z
task/TASK-43      ghi9012         2026-03-10T11:15:00Z
```

#### lango p2p git branch merge

Merge a task branch into a target branch. Uses `git merge-tree --write-tree` for conflict detection in bare repositories without needing a working tree.

```
lango p2p git branch merge <workspace-id> --task <task-id> [--into <branch>]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--task` | string | *required* | Task ID of the source branch (`task/{taskID}`) |
| `--into` | string | `main` | Target branch to merge into |

If conflicts are detected, the merge is aborted and conflicting file paths are reported.

**Example (success):**

```bash
$ lango p2p git branch merge a1b2c3d4-... --task TASK-42
Merged task/TASK-42 into main (commit: abc1234)
```

**Example (conflict):**

```bash
$ lango p2p git branch merge a1b2c3d4-... --task TASK-43
Merge conflict between task/TASK-43 and main:
  - src/optimizer.go
  - src/config.go
```

#### lango p2p git branch delete

Delete a task branch from the workspace repository.

```
lango p2p git branch delete <workspace-id> --task <task-id>
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--task` | string | *required* | Task ID of the branch to delete |

The operation is idempotent — if the branch does not exist, it succeeds without error.

**Example:**

```bash
$ lango p2p git branch delete a1b2c3d4-... --task TASK-42
Deleted branch task/TASK-42 from workspace a1b2c3d4-...
```

---

### Workflow Example: Task Branch Lifecycle

A typical collaboration workflow using task branches and incremental bundles:

```bash
# 1. Initialize workspace and create a task branch
lango p2p git init a1b2c3d4-...
lango p2p git branch create a1b2c3d4-... --task TASK-42

# 2. Work on the task branch (commits are made via agent tools)
#    ... agent writes code and commits to task/TASK-42 ...

# 3. Push an incremental bundle to peers (only new commits since last sync)
lango p2p git push a1b2c3d4-...

# 4. Peers fetch and safely apply the bundle
lango p2p git fetch a1b2c3d4-...

# 5. Merge the completed task branch into main
lango p2p git branch merge a1b2c3d4-... --task TASK-42

# 6. Clean up the task branch
lango p2p git branch delete a1b2c3d4-... --task TASK-42
```

### Git Bundle Features

- **Bare Repositories**: Each workspace has an isolated bare git repo at `~/.lango/workspaces/<id>/repo.git`
- **Bundle Protocol**: Uses `git bundle create/unbundle` for atomic transfers over the P2P network
- **Incremental Bundles**: Transfer only new commits since a known base, with automatic full-bundle fallback
- **Safe Apply**: Verify prerequisites, snapshot refs, apply, and auto-rollback on failure
- **Task Branches**: Per-task isolation via `task/{taskID}` branches with idempotent create/delete
- **Conflict Detection**: `git merge-tree --write-tree` detects merge conflicts without a working tree
- **DAG Leaf Detection**: Identifies leaf commits (no children) for conflict detection
- **Size Limits**: Configurable `maxBundleSizeBytes` to prevent oversized transfers

Git operations require the `git` binary to be installed and available in PATH.
