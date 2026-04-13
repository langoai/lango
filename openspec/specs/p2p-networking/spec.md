## Purpose

Capability spec for p2p-networking. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: libp2p Node Lifecycle

The P2P `Node` SHALL encapsulate a libp2p host with an Ed25519 identity key. When `*security.SecretsStore` is provided, the key SHALL be stored encrypted in SecretsStore under `p2p.node.privatekey`. When SecretsStore is nil, the key SHALL be persisted at `{keyDir}/node.key` for backward compatibility. The node key SHALL be loaded on startup (priority: SecretsStore → legacy file → generate new), ensuring peer identity survives restarts. The node MUST use Noise protocol encryption on all connections.

#### Scenario: Node key persists across restarts
- **WHEN** a `Node` is created and a node key exists (in SecretsStore or as `node.key`)
- **THEN** the node SHALL load the existing key and present the same peer ID as the previous instance

#### Scenario: Node key generated on first start
- **WHEN** a `Node` is created and no node key exists
- **THEN** the node SHALL generate a new Ed25519 keypair, persist it (to SecretsStore if available, else to `node.key` with `0600`), and use it as the peer identity

#### Scenario: Node creation with invalid keyDir
- **WHEN** `NewNode` is called with a `keyDir` path that cannot be created and SecretsStore is nil
- **THEN** `NewNode` SHALL return an error and SHALL NOT start any host or network listener

---

### Requirement: Node constructor accepts SecretsStore
`NewNode()` SHALL accept an optional `*security.SecretsStore` parameter for encrypted node key management. When nil, file-based storage is used.

#### Scenario: Node created with SecretsStore
- **WHEN** `NewNode(cfg, logger, secrets)` is called with a non-nil SecretsStore
- **THEN** the node SHALL use SecretsStore for key storage

#### Scenario: Node created without SecretsStore
- **WHEN** `NewNode(cfg, logger, nil)` is called
- **THEN** the node SHALL fall back to file-based key storage in `cfg.KeyDir`

---

### Requirement: Kademlia DHT Bootstrap

The `Node.Start` method SHALL initialize a Kademlia DHT in `ModeAutoServer` and call `Bootstrap` to enter the DHT routing table. The node SHALL attempt to connect to each configured bootstrap peer concurrently using goroutines bounded by the caller-provided `sync.WaitGroup`. Bootstrap peer connection failures MUST be logged as warnings and SHALL NOT prevent the node from starting.

#### Scenario: Successful DHT bootstrap with bootstrap peers
- **WHEN** `Node.Start` is called with one or more valid bootstrap peer multiaddrs
- **THEN** the node SHALL connect to each bootstrap peer and log "connected to bootstrap peer"

#### Scenario: Invalid bootstrap peer address
- **WHEN** a configured bootstrap peer address is not a valid multiaddr
- **THEN** the node SHALL log a warning with the invalid address and SHALL continue starting with the remaining peers

#### Scenario: DHT bootstrap failure
- **WHEN** `dht.Bootstrap` returns an error
- **THEN** `Node.Start` SHALL call the context cancel function, close the DHT, and return a wrapped error

---

### Requirement: mDNS LAN Discovery

When `cfg.EnableMDNS` is true, the `Node.Start` method SHALL start an mDNS service using the libp2p `mdns.NewMdnsService`. The mDNS notifee SHALL automatically connect to discovered LAN peers. The node's own peer ID SHALL be excluded from connection attempts. mDNS startup failures MUST be logged as warnings and SHALL NOT prevent the node from completing startup.

#### Scenario: mDNS peer discovery and auto-connect
- **WHEN** a peer on the same LAN broadcasts its presence via mDNS
- **THEN** the local node SHALL call `host.Connect` with the discovered peer info and log "mDNS peer discovered"

#### Scenario: mDNS discovers own peer ID
- **WHEN** the mDNS service receives a discovery event for the local node's own peer ID
- **THEN** the notifee SHALL silently ignore the event and SHALL NOT attempt to connect to itself

---

### Requirement: Connection Manager Watermarks

The `Node` SHALL create a `connmgr.ConnManager` with `maxPeers` as the high watermark and `maxPeers * 80 / 100` as the low watermark. The connection manager MUST trim excess connections when the high watermark is reached, pruning down to the low watermark.

#### Scenario: Connections pruned at high watermark
- **WHEN** the number of connected peers reaches `cfg.MaxPeers`
- **THEN** the connection manager SHALL trim the least-recently-used connections until the peer count reaches the low watermark

#### Scenario: Zero maxPeers rejected
- **WHEN** `connmgr.NewConnManager` is called with a zero or negative high watermark
- **THEN** `NewNode` SHALL return an error from the connection manager initialization

---

### Requirement: Graceful Shutdown

`Node.Stop` SHALL cancel the internal context, close the mDNS service (if started), close the DHT, and close the libp2p host in that order. Any error from DHT or host close SHALL be returned. mDNS close errors MUST be logged as warnings and SHALL NOT prevent further shutdown steps.

#### Scenario: Clean stop sequence
- **WHEN** `Node.Stop` is called on a running node
- **THEN** the node SHALL cancel its context, close mDNS, close the DHT, close the host, and log "P2P node stopped"

#### Scenario: Stop on partially initialized node
- **WHEN** `Node.Stop` is called on a node where `Start` was not called
- **THEN** `Node.Stop` SHALL return nil without panicking (nil checks on `cancel`, `mdnsSvc`, and `dht`)

---

### Requirement: Protocol Stream Handler Registration

The `Node.SetStreamHandler` method SHALL register a `network.StreamHandler` for the given protocol ID on the underlying libp2p host. The `Node.Host()` method SHALL expose the underlying `host.Host` for direct protocol registration by sub-packages.

#### Scenario: Stream handler registration
- **WHEN** `Node.SetStreamHandler("/lango/a2a/1.0.0", handler)` is called
- **THEN** all incoming streams with protocol `/lango/a2a/1.0.0` SHALL be dispatched to `handler`
