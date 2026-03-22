## ADDED Requirements

### Requirement: P2P package uses ConfigReader interface instead of config.P2PConfig
`internal/p2p/node.go` SHALL accept a `ConfigReader` interface instead of `config.P2PConfig` directly. The `internal/p2p/` root package MUST NOT import `internal/config/`.

#### Scenario: NewNode accepts ConfigReader
- **WHEN** `p2p.NewNode(cfg, logger, secrets)` is called
- **THEN** `cfg` parameter type is `p2p.ConfigReader`, not `config.P2PConfig`

#### Scenario: P2P package has no config import
- **WHEN** checking imports of `internal/p2p/node.go`
- **THEN** `internal/config` is not in the import list

### Requirement: config.P2PConfig satisfies p2p.ConfigReader
`config.P2PConfig` SHALL have value-receiver getter methods that implicitly satisfy `p2p.ConfigReader`. No explicit adapter type is needed.

#### Scenario: P2PConfig passes to NewNode without wrapper
- **WHEN** app wiring calls `p2p.NewNode(cfg.P2P, ...)`
- **THEN** it compiles without an adapter because P2PConfig satisfies ConfigReader

### Requirement: ConfigReader covers all node.go field accesses
The `ConfigReader` interface SHALL expose methods for every config field that `node.go` accesses: `GetKeyDir()`, `GetMaxPeers()`, `GetListenAddrs()`, `GetEnableRelay()`, `GetBootstrapPeers()`, `GetEnableMDNS()`.

#### Scenario: All node.go config accesses use getter methods
- **WHEN** `node.go` needs a config value
- **THEN** it calls a `ConfigReader` method, not a struct field access
