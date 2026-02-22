## MODIFIED Requirements

### Requirement: Agent Card Extended with P2P Fields

The `AgentCard` struct (served at `GET /.well-known/agent.json`) SHALL be extended with the following optional P2P fields in addition to its existing `name`, `description`, `url`, and `skills` fields:

- `did` (`string`, omitempty): The agent's decentralized identifier in `did:lango:<hex-pubkey>` format, populated when P2P is enabled.
- `multiaddrs` (`[]string`, omitempty): The list of libp2p multiaddresses the agent is reachable at over the P2P network.
- `capabilities` (`[]string`, omitempty): A list of capability identifiers the agent advertises for P2P capability-based discovery.
- `pricing` (`*PricingInfo`, omitempty): Optional pricing structure containing `currency`, `perQuery`, `perMinute`, and `toolPrices` map. Currency SHALL be `"USDC"`.
- `zkCredentials` (`[]ZKCredential`, omitempty): Optional list of ZK-attested capability credentials, each containing `capabilityId`, `proof` (bytes), `issuedAt`, and `expiresAt`.

When P2P is disabled, all P2P extension fields SHALL be omitted from the JSON output (via `omitempty`). The HTTP endpoint behavior, path, and content-type SHALL remain unchanged.

#### Scenario: Agent card includes P2P fields when P2P enabled
- **WHEN** `GET /.well-known/agent.json` is called and P2P is enabled with a DID and multiaddrs configured
- **THEN** the response JSON SHALL include `did`, `multiaddrs`, and `capabilities` fields with their configured values

#### Scenario: P2P fields absent when P2P disabled
- **WHEN** `GET /.well-known/agent.json` is called and P2P is disabled
- **THEN** the response JSON SHALL NOT contain `did`, `multiaddrs`, `capabilities`, `pricing`, or `zkCredentials` fields

#### Scenario: SetP2PInfo populates card fields
- **WHEN** `Server.SetP2PInfo(did, multiaddrs, capabilities)` is called on an A2A server
- **THEN** subsequent calls to `GET /.well-known/agent.json` SHALL return the provided DID, multiaddrs, and capabilities

#### Scenario: Pricing info serialized correctly
- **WHEN** `Server.SetPricing(&PricingInfo{Currency: "USDC", PerQuery: "0.01"})` is called
- **THEN** the agent card JSON SHALL contain `"pricing": {"currency": "USDC", "perQuery": "0.01"}`

#### Scenario: ZK credentials included in agent card
- **WHEN** `AgentCard.ZKCredentials` contains a credential with a non-expired `ExpiresAt`
- **THEN** the credential SHALL appear in the JSON output with all fields present

---

### Requirement: Agent Card Served Without Authentication

The `GET /.well-known/agent.json` endpoint SHALL remain publicly accessible without any authentication requirement. P2P extension fields in the card (DID, multiaddrs) are intentionally public information used for peer discovery and SHALL be served to any requester.

#### Scenario: Unauthenticated request receives full agent card
- **WHEN** an unauthenticated HTTP GET is made to `/.well-known/agent.json`
- **THEN** the server SHALL respond with HTTP 200 and the full agent card JSON including any P2P extension fields

---

### Requirement: GossipCard Mirrors AgentCard P2P Fields

The `GossipCard` type used for GossipSub propagation SHALL carry the same P2P-related fields as the `AgentCard` extension: `name`, `description`, `did`, `multiaddrs`, `capabilities`, `pricing`, `zkCredentials`, `peerId`, and `timestamp`. The `GossipCard` is separate from `AgentCard` but SHALL be structurally consistent with the P2P extension fields to enable seamless conversion between the two representations.

#### Scenario: GossipCard fields match AgentCard P2P fields
- **WHEN** a `GossipCard` is constructed from an `AgentCard` with P2P fields set
- **THEN** all P2P extension fields (`did`, `multiaddrs`, `capabilities`, `pricing`, `zkCredentials`) SHALL be preserved in the `GossipCard`
