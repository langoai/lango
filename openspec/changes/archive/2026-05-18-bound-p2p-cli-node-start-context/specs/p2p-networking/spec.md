## MODIFIED Requirements

### Requirement: Kademlia DHT Bootstrap
The `Node.Start` method SHALL initialize a Kademlia DHT in `ModeAutoServer` and call `Bootstrap` to enter the DHT routing table. The node SHALL attempt to connect to each configured bootstrap peer concurrently using goroutines bounded by the caller-provided `sync.WaitGroup`. Bootstrap peer connection failures MUST be logged as warnings and SHALL NOT prevent the node from starting.

#### Scenario: Node startup derives lifecycle from caller context
- **WHEN** `Node.Start` is called with a parent context
- **THEN** DHT bootstrap, bootstrap peer connection attempts, and mDNS notifee connection attempts SHALL use a node lifecycle context derived from that parent
- **AND** canceling the parent context SHALL cancel in-flight startup work without requiring `context.Background()`
