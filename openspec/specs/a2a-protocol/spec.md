## ADDED Requirements

### Requirement: Agent Card endpoint
The system SHALL serve an Agent Card at `GET /.well-known/agent.json` when A2A is enabled, containing the agent's name, description, URL, and skills.

The `AgentCard` struct SHALL be extended with the following optional P2P fields in addition to its existing `name`, `description`, `url`, and `skills` fields:

- `did` (`string`, omitempty): The agent's decentralized identifier in `did:lango:<hex-pubkey>` format, populated when P2P is enabled.
- `multiaddrs` (`[]string`, omitempty): The list of libp2p multiaddresses the agent is reachable at over the P2P network.
- `capabilities` (`[]string`, omitempty): A list of capability identifiers the agent advertises for P2P capability-based discovery.
- `pricing` (`*PricingInfo`, omitempty): Optional pricing structure containing `currency`, `perQuery`, `perMinute`, and `toolPrices` map. Currency SHALL be `"USDC"`.
- `zkCredentials` (`[]ZKCredential`, omitempty): Optional list of ZK-attested capability credentials, each containing `capabilityId`, `proof` (bytes), `issuedAt`, and `expiresAt`.

When P2P is disabled, all P2P extension fields SHALL be omitted from the JSON output (via `omitempty`). The HTTP endpoint behavior, path, and content-type SHALL remain unchanged.

#### Scenario: Agent card served
- **WHEN** a GET request is made to `/.well-known/agent.json`
- **THEN** the response SHALL be JSON with `name`, `description`, `url`, and `skills` fields

#### Scenario: Skills derived from agent tree
- **WHEN** the agent has sub-agents (multi-agent mode)
- **THEN** each sub-agent SHALL appear as a skill in the Agent Card

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

---

### Requirement: A2A server route mounting
The A2A server SHALL mount its routes on the gateway's chi.Router when `a2a.enabled` and `agent.multiAgent` are both true.

#### Scenario: Routes mounted on gateway
- **WHEN** the application starts with both A2A and multi-agent enabled
- **THEN** the A2A server's RegisterRoutes SHALL be called with the gateway's Router

#### Scenario: A2A disabled
- **WHEN** `a2a.enabled` is false
- **THEN** no A2A routes SHALL be mounted

### Requirement: Gateway Router accessor
The gateway Server SHALL expose a `Router() chi.Router` method for external route mounting.

#### Scenario: Router method returns chi.Router
- **WHEN** `Router()` is called on the gateway server
- **THEN** it SHALL return the internal chi.Router instance

### Requirement: ADK Agent accessor
The adk.Agent SHALL expose an `ADKAgent()` method returning the underlying `adk_agent.Agent` for use by A2A server.

#### Scenario: ADKAgent returns underlying agent
- **WHEN** `ADKAgent()` is called on an adk.Agent created via NewAgent or NewAgentFromADK
- **THEN** it SHALL return the stored adk_agent.Agent instance

### Requirement: Remote Agent Loading Order
Remote A2A agents SHALL be loaded and assigned to `orchCfg.RemoteAgents` BEFORE calling `BuildAgentTree()`, ensuring they are included in the orchestrator's sub-agent list.

#### Scenario: A2A agents configured
- **WHEN** `cfg.A2A.Enabled` is true and remote agents are configured
- **THEN** remote agents SHALL be loaded and available in `orchCfg.RemoteAgents` before `BuildAgentTree()` is called

#### Scenario: A2A loading fails
- **WHEN** remote agent loading produces an error
- **THEN** the error SHALL be logged as a warning and the agent tree SHALL still be built without remote agents
