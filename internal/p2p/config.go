package p2p

// ConfigReader is a narrow interface for the configuration fields that Node
// needs. Consumers define the interface (Go convention) so the p2p package
// does not depend on the config package.
type ConfigReader interface {
	// GetKeyDir returns the legacy directory for persisting node keys.
	GetKeyDir() string

	// GetMaxPeers returns the maximum number of connected peers.
	GetMaxPeers() int

	// GetListenAddrs returns the multiaddrs to listen on.
	GetListenAddrs() []string

	// GetEnableRelay reports whether this node acts as a relay for NAT traversal.
	GetEnableRelay() bool

	// GetBootstrapPeers returns the initial peers for DHT bootstrapping.
	GetBootstrapPeers() []string

	// GetEnableMDNS reports whether multicast DNS discovery is enabled.
	GetEnableMDNS() bool
}
