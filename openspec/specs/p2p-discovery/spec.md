## ADDED Requirements

### Requirement: GossipSub Agent Card Propagation

The `GossipService` SHALL join the GossipSub topic `/lango/agentcard/1.0.0` and periodically publish the local `GossipCard` at the configured interval. The card SHALL be published immediately on service start. Own messages SHALL be discarded (filtered by comparing `msg.ReceivedFrom` to `host.ID()`). The publisher and subscriber SHALL run in separate goroutines tracked by a `sync.WaitGroup`.

#### Scenario: Card published immediately on start
- **WHEN** `GossipService.Start(wg)` is called
- **THEN** the local agent card SHALL be published to the topic within the first tick cycle (immediately)

#### Scenario: Card published periodically
- **WHEN** `GossipService.Start` is called with `Interval=30s`
- **THEN** the card SHALL be re-published every 30 seconds with an updated `Timestamp`

#### Scenario: Own messages ignored
- **WHEN** the GossipSub subscription delivers a message whose `ReceivedFrom` equals the local host ID
- **THEN** the `subscribeLoop` SHALL discard the message without updating the peer map

#### Scenario: Nil local card skips publication
- **WHEN** `GossipService` is initialized with a nil `LocalCard`
- **THEN** `publishCard` SHALL return immediately without encoding or publishing

---

### Requirement: ZK Credential Verification on Received Cards

When a `GossipCard` is received containing `ZKCredentials`, the `GossipService` SHALL verify each non-expired credential using the configured `ZKCredentialVerifier`. If any credential fails verification, the entire card MUST be discarded. Expired credentials SHALL be skipped (logged at debug level) and SHALL NOT cause the card to be discarded.

#### Scenario: Card with valid ZK credentials stored
- **WHEN** a received `GossipCard` has one ZK credential that passes `ZKCredentialVerifier`
- **THEN** the card SHALL be stored in the peer map under its DID

#### Scenario: Card with invalid ZK credential discarded
- **WHEN** a received `GossipCard` has a ZK credential for which the `ZKCredentialVerifier` returns `(false, nil)` or an error
- **THEN** the card SHALL NOT be stored and the discardal SHALL be logged as a warning

#### Scenario: Card with expired credential not discarded for that credential
- **WHEN** a received `GossipCard` has a ZK credential whose `ExpiresAt` is before `time.Now()`
- **THEN** that credential SHALL be skipped (debug log) and the card SHALL still be accepted if all other credentials are valid

---

### Requirement: Peer Card Deduplication by Timestamp

The `GossipService` SHALL update the peer map only when the incoming card's `Timestamp` is strictly after the stored card's `Timestamp`. If the incoming card is older or equal in timestamp, it SHALL be silently discarded. Cards with an empty `DID` field MUST be discarded unconditionally.

#### Scenario: Newer card replaces older card
- **WHEN** a card with a newer `Timestamp` arrives for an already-known DID
- **THEN** the peer map SHALL be updated with the new card

#### Scenario: Older card not stored
- **WHEN** a card with a `Timestamp` older than the stored card arrives for the same DID
- **THEN** the peer map SHALL retain the existing card

#### Scenario: Card with empty DID discarded
- **WHEN** a received `GossipCard` has `DID: ""`
- **THEN** `handleMessage` SHALL return immediately without storing the card

---

### Requirement: Capability and DID Lookup on Known Peers

`GossipService.FindByCapability` SHALL return all stored `GossipCard` entries that list the requested capability string in their `Capabilities` slice. `GossipService.FindByDID` SHALL return the stored card for an exact DID match, or nil if not found. `GossipService.KnownPeers` SHALL return a snapshot of all stored cards.

#### Scenario: Capability search returns matching peers
- **WHEN** `FindByCapability("code_execution")` is called and two peers advertise that capability
- **THEN** both cards SHALL be returned

#### Scenario: DID lookup returns exact match
- **WHEN** `FindByDID("did:lango:abc")` is called and the DID is in the peer map
- **THEN** the corresponding `GossipCard` SHALL be returned

#### Scenario: DID lookup returns nil for unknown DID
- **WHEN** `FindByDID("did:lango:unknown")` is called
- **THEN** nil SHALL be returned

---

### Requirement: DHT Agent Advertisement

The `AdService` SHALL publish the local `AgentAd` to the Kademlia DHT under the key `/lango/agentad/<did>` using `dht.PutValue`. `AdService.Discover` SHALL filter stored `AgentAd` entries by tag match (any tag matches). `AdService.StoreAd` SHALL verify ZK credentials before storing and MUST reject ads with empty DIDs.

#### Scenario: Agent ad published to DHT
- **WHEN** `AdService.Advertise(ctx)` is called
- **THEN** the local `AgentAd` SHALL be JSON-marshaled and stored in the DHT under `/lango/agentad/<localDID>`

#### Scenario: Discovery by tag returns matching ads
- **WHEN** `AdService.Discover(ctx, []string{"researcher"})` is called and one stored ad has tag `"researcher"`
- **THEN** only that ad SHALL be returned

#### Scenario: Discover with no tags returns all ads
- **WHEN** `AdService.Discover(ctx, nil)` is called
- **THEN** all stored ads SHALL be returned

#### Scenario: Ad with invalid ZK credential rejected on store
- **WHEN** `StoreAd` is called with an ad containing a ZK credential that fails verification
- **THEN** `StoreAd` SHALL return an error and SHALL NOT store the ad

#### Scenario: Ad with empty DID rejected
- **WHEN** `StoreAd` is called with an ad where `DID == ""`
- **THEN** `StoreAd` SHALL return an error containing "agent ad missing DID"
