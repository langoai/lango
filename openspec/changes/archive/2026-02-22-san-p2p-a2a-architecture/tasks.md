## 1. Config Extension

- [x] 1.1 Add P2PConfig struct to internal/config/types.go (Enabled, ListenAddrs, BootstrapPeers, KeyDir, EnableRelay, EnableMDNS, MaxPeers, HandshakeTimeout, SessionTokenTTL, AutoApproveKnownPeers, FirewallRules, GossipInterval, ZKHandshake, ZKAttestation)
- [x] 1.2 Add P2P field to root Config struct
- [x] 1.3 Add default config values and validation in loader.go

## 2. Dependencies

- [x] 2.1 Install go-libp2p v0.47.0, go-libp2p-kad-dht, go-libp2p-pubsub, go-multiaddr
- [x] 2.2 Install gnark v0.14.0 for ZKP circuits
- [x] 2.3 Run go mod tidy

## 3. Wallet Extension

- [x] 3.1 Add PublicKey(ctx) ([]byte, error) to WalletProvider interface
- [x] 3.2 Implement PublicKey in LocalWallet (crypto.CompressPubkey)
- [x] 3.3 Implement PublicKey delegation in RPCWallet and CompositeWallet

## 4. P2P Node

- [x] 4.1 Create internal/p2p/node.go with libp2p host (Noise, TCP/QUIC, ConnManager)
- [x] 4.2 Implement loadOrGenerateKey for Ed25519 node key persistence
- [x] 4.3 Implement Start() with DHT bootstrap, mDNS discovery, bootstrap peer connection
- [x] 4.4 Implement Stop() with graceful shutdown (mDNS -> DHT -> host)

## 5. Identity/DID

- [x] 5.1 Create internal/p2p/identity/identity.go with DID struct and Provider interface
- [x] 5.2 Implement WalletDIDProvider (did:lango:<hex> from wallet public key)
- [x] 5.3 Implement ParseDID and DIDFromPublicKey helper functions
- [x] 5.4 Implement VerifyDID for peer identity verification

## 6. ZKP Core

- [x] 6.1 Create internal/zkp/zkp.go with ProverService (PlonK/Groth16 on BN254)
- [x] 6.2 Implement Compile, Prove, Verify with gnark backend
- [x] 6.3 Create WalletOwnershipCircuit (circuits/ownership.go)
- [x] 6.4 Create BalanceRangeCircuit (circuits/balance.go)
- [x] 6.5 Create ResponseAttestationCircuit (circuits/attestation.go)
- [x] 6.6 Create AgentCapabilityCircuit (circuits/capability.go)

## 7. Handshake

- [x] 7.1 Create internal/p2p/handshake/handshake.go with Handshaker (Challenge-Response-Ack)
- [x] 7.2 Implement ZK-enhanced mode with ECDSA signature fallback
- [x] 7.3 Implement HITL approval callback pattern (ApprovalFunc)
- [x] 7.4 Create internal/p2p/handshake/session.go with HMAC-SHA256 token store and TTL eviction

## 8. Knowledge Firewall

- [x] 8.1 Create internal/p2p/firewall/firewall.go with default deny-all ACL
- [x] 8.2 Implement FilterQuery with per-peer rate limiting
- [x] 8.3 Implement SanitizeResponse for sensitive field removal
- [x] 8.4 Implement AttestResponse for ZK attestation callback
- [x] 8.5 Implement dynamic rule Add/Remove operations

## 9. A2A-over-P2P Protocol

- [x] 9.1 Create internal/p2p/protocol/messages.go with Request/Response types
- [x] 9.2 Create internal/p2p/protocol/handler.go with session validation + firewall + attestation
- [x] 9.3 Create internal/p2p/protocol/remote_agent.go as P2P remote agent adapter
- [x] 9.4 Implement SendRequest utility for client-side stream communication

## 10. Discovery

- [x] 10.1 Create internal/p2p/discovery/gossip.go with GossipSub agent card propagation
- [x] 10.2 Implement ZK credential verification on received cards
- [x] 10.3 Implement FindByCapability and FindByDID peer lookups
- [x] 10.4 Create internal/p2p/discovery/agentad.go with DHT-based agent advertisements

## 11. Agent Card P2P Extension

- [x] 11.1 Add DID, Multiaddrs, Capabilities, Pricing, ZKCredentials to AgentCard in a2a/server.go
- [x] 11.2 Add PricingInfo and ZKCredential types
- [x] 11.3 Add SetP2PInfo and SetPricing methods to A2A Server
- [x] 11.4 Add Card() accessor method

## 12. App Wiring

- [x] 12.1 Add P2PNode field to App struct in types.go
- [x] 12.2 Create initP2P() in wiring.go (node, identity, sessions, handshaker, firewall, protocol handler, gossip)
- [x] 12.3 Wire P2P Start/Stop in app.go lifecycle
- [x] 12.4 Register handshake and A2A protocol handlers on libp2p host

## 13. P2P Tools

- [x] 13.1 Implement p2p_status tool (Safe)
- [x] 13.2 Implement p2p_connect tool (Dangerous) with handshake
- [x] 13.3 Implement p2p_disconnect tool (Moderate)
- [x] 13.4 Implement p2p_peers tool (Safe)
- [x] 13.5 Implement p2p_query tool (Moderate) with remote agent adapter
- [x] 13.6 Implement p2p_firewall_rules/add/remove tools
- [x] 13.7 Implement p2p_discover tool (Safe)

## 14. P2P Payment

- [x] 14.1 Implement p2p_pay tool (Dangerous) using payment.Service.Send
- [x] 14.2 Wire session verification and DID-to-address derivation

## 15. Orchestration

- [x] 15.1 Add "p2p_" prefix to vault agent Prefixes in orchestration/tools.go
- [x] 15.2 Add P2P-related keywords to vault agent Keywords

## 16. Build Verification

- [x] 16.1 Run go build ./... — all packages compile
- [x] 16.2 Run go test ./... — all existing tests pass
