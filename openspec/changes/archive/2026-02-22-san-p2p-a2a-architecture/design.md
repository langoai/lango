## Context

Lango is a Go AI agent platform with HTTP-based A2A communication, ECDSA wallet (USDC on Base), and a security layer (CryptoProvider/SecretsStore). All inter-agent communication currently depends on a central HTTP server, making discovery and communication registry-dependent. This change adds a decentralized P2P networking layer — allowing agents to discover each other, mutually authenticate, and exchange A2A messages without any central coordinator, while preserving the existing wallet and security infrastructure.

## Goals / Non-Goals

### Goals

- Platform-independent agent discovery via Kademlia DHT and mDNS (LAN fallback)
- Zero-trust mutual authentication using ZK-enhanced handshakes derived from existing wallet keys
- Knowledge privacy enforcement via a default deny-all firewall on all incoming P2P queries
- User sovereignty (HITL): agent-to-agent interactions require explicit user approval during handshake
- Peer-to-peer USDC payments over P2P streams using existing payment service
- Zero new key management: DID identity is derived from the existing wallet ECDSA public key

### Non-Goals

- Production MPC ceremony for gnark SRS parameters (trusted setup)
- Smart contract deployment or on-chain DID registry
- Mobile or browser P2P support
- GUI/TUI for real-time P2P management
- Per-message ZKP verification (proof generation latency is acceptable only at handshake time)

## Decisions

### 1. libp2p over Custom Networking

**Options considered**:
- Custom TCP/TLS with self-signed certs
- WebRTC with STUN/TURN
- libp2p (go-libp2p)

**Decision**: libp2p v0.47.0

libp2p provides Noise protocol encryption, TCP and QUIC transports, Kademlia DHT, mDNS discovery, and GossipSub pub/sub — all battle-tested in IPFS and Filecoin production networks. Building equivalent functionality from scratch would introduce significant security surface. libp2p's `peer.ID` maps naturally to a content-addressed identity, and its stream multiplexing integrates cleanly with the A2A request/response pattern.

### 2. DID Derived from Wallet Key

**Options considered**:
- Separate Ed25519 keypair for P2P identity
- did:key method with new key generation
- did:lango derived from existing ECDSA wallet public key

**Decision**: `did:lango:<compressed-secp256k1-pubkey-hex>`

The existing wallet (`payment.enabled`) already holds an ECDSA keypair used for on-chain transactions. Deriving the DID from the same public key ties P2P identity to on-chain identity at zero operational cost — no additional key generation, rotation policies, or backup procedures. The `PublicKey()` method was added to the `WalletProvider` interface to surface the compressed public key. P2P is gated on `payment.enabled`; agents without a wallet cannot participate in the P2P network.

### 3. ZKP: gnark Circuits with Hash-Based Fallback

**Options considered**:
- Pure signature-based authentication (no ZKP)
- External ZKP service (snarkjs/rapidsnark via subprocess)
- gnark native Go circuits (PlonK on BN254)

**Decision**: gnark v0.14.0 with hash-based development fallback

Four circuits are defined: `OwnershipCircuit` (proves control of DID private key), `BalanceCircuit` (proves USDC balance above threshold without revealing amount), `AttestationCircuit` (proves possession of a signed credential), and `CapabilityCircuit` (proves agent capability without revealing implementation). PlonK on BN254 is used for its universal trusted setup (no per-circuit ceremony). A hash-based placeholder (`zkp.HashProver`) is provided for development and testing environments where gnark's trusted setup is unavailable. The `ZKProverFunc` callback in `HandshakeConfig` allows injection of either implementation.

**Trade-off**: gnark adds approximately 6-8 MB to the binary. This is acceptable for a server-side agent platform. Proof generation takes 50-200ms per proof, which is acceptable at handshake time but not per-message.

### 4. Callback Pattern for Import Cycle Avoidance

**Options considered**:
- Direct interface imports between `internal/p2p/` and `internal/app/`
- Separate adapter package
- Callback functions injected at wiring time

**Decision**: Callback functions injected at wiring time

This matches the existing `EmbedCallback`/`GraphCallback` pattern established in the codebase. Four callback types are defined on P2P config structures: `ToolExecutor` (executes agent tools on behalf of remote peers), `CardProvider` (returns the local agent's A2A card), `ApprovalFunc` (HITL: blocks handshake until user approves), and `ZKProverFunc` (generates ZK proofs). All are wired in `internal/app/wiring.go`. The `internal/p2p/` package has no import dependency on `internal/app/`.

### 5. Default Deny-All Knowledge Firewall

**Decision**: `KnowledgeFirewall` blocks all incoming P2P queries by default. Explicit allow rules are required per-capability, per-peer, or per-DID pattern.

Zero-trust by design: an agent joining the P2P network does not automatically share any knowledge. Operators must explicitly configure which capabilities remote peers may invoke and from which DIDs. Rules are evaluated in order; the first match wins. A catch-all deny rule is always appended as the final rule. Response sanitization strips fields matching configured patterns before returning results to remote peers.

### 6. Session-Based Auth with HMAC-SHA256 Tokens

**Decision**: After a successful handshake (wallet signature verification + optional ZKP verification + HITL approval), a session token is issued using HMAC-SHA256 over `(peerID + sessionID + timestamp)` with a configurable TTL (default 24h). Subsequent A2A messages over P2P streams present this token to skip the full handshake overhead.

Session state is held in memory (map protected by `sync.RWMutex`) on each node. Sessions are not persisted across restarts — reconnection triggers a new handshake. This is intentional: it keeps the session store simple and avoids persistent storage dependencies in the P2P layer.

### 7. GossipSub for Agent Card Propagation

**Decision**: Agent cards (extended with DID, multiaddrs, capabilities, pricing, and ZK credentials) are broadcast periodically over a GossipSub topic (`lango/agent-cards/v1`). Received cards are verified against the sender's DID before being indexed for capability search.

**Trade-off**: GossipSub fans out messages to all subscribers, which can amplify traffic in large networks. Mitigation: per-peer rate limiting on card reception (max 1 card/minute per peer), and card deduplication by content hash. DHT advertisements (`dht.Provide`) are used in parallel for targeted capability lookup without broadcast.

### 8. ConnManager with High/Low Watermarks

**Decision**: `connmgr.NewConnManager` with configurable `maxPeers` (default 50), low watermark at 80% of max (40), and graceful trim on excess. This prevents unbounded peer accumulation while maintaining a healthy routing table for DHT.

### 9. Vault Agent Routes p2p_ Tools

**Decision**: The 10 P2P tools (`p2p_status`, `p2p_connect`, `p2p_disconnect`, `p2p_peers`, `p2p_query`, `p2p_firewall_rules`, `p2p_firewall_add`, `p2p_firewall_remove`, `p2p_discover`, `p2p_pay`) are routed through the vault agent, consistent with the existing pattern for security-sensitive tools (`crypto_*`, `secrets_*`, `payment_*`). This centralizes privileged tool routing without requiring a separate P2P agent role.

## Risks / Trade-offs

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| gnark binary size increase (+6-8 MB) | Certain | Low | Acceptable for server-side platform; document in release notes |
| GossipSub message amplification in large networks | Medium | Medium | Per-peer rate limiting (1 card/min), content-hash deduplication |
| DHT bootstrap cold start (no known peers) | High | Medium | mDNS as automatic LAN fallback; configurable bootstrap peer list |
| ZKP proof generation latency (50-200ms) | Certain | Low | ZKP only at handshake; session tokens amortize cost for subsequent messages |
| gnark trusted setup (SRS) in production | Medium | High | Hash-based fallback for dev; document MPC ceremony requirement for production |
| In-memory session state lost on restart | Certain | Low | Intentional design; reconnect triggers new handshake; document behavior |
| P2P port exposure in Docker/firewall | Medium | Medium | Configurable listen addresses; document required port (default 4001/tcp+udp) |
| HITL approval blocking async P2P queries | Low | Medium | Approval timeout configurable; timeout → auto-deny to prevent hangs |
