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
