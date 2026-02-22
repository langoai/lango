## Why

Lango currently relies on centralized HTTP-based A2A communication. To achieve platform-independent, censorship-resistant agent networking — like IPFS/Bitcoin for AI agents — we need a decentralized P2P layer where agents can discover, authenticate, communicate, and transact without central registries.

## What Changes

- Add libp2p-based P2P networking node with Kademlia DHT and mDNS discovery
- Introduce DID identity system derived from existing wallet public keys (`did:lango:<pubkey>`)
- Implement ZK-enhanced peer authentication (handshake with wallet signatures + optional ZKP)
- Add Knowledge Firewall with default deny-all ACL and response sanitization
- Implement A2A-over-P2P protocol for tool invocation over encrypted libp2p streams
- Add GossipSub-based agent card propagation and DHT-based agent advertisements
- Extend Agent Card with P2P fields (DID, multiaddrs, capabilities, pricing, ZK credentials)
- Add ZKP core using gnark (PlonK/Groth16) for ownership, balance, attestation, and capability circuits
- Add peer-to-peer USDC payment via existing payment service
- Add 11 new agent tools (`p2p_status`, `p2p_connect`, `p2p_disconnect`, `p2p_peers`, `p2p_query`, `p2p_firewall_rules`, `p2p_firewall_add`, `p2p_firewall_remove`, `p2p_discover`, `p2p_pay`)

## Capabilities

### New Capabilities
- `p2p-networking`: libp2p node lifecycle, DHT bootstrap, mDNS discovery, peer connection management
- `p2p-identity`: DID derivation from wallet, peer identity verification
- `p2p-handshake`: ZK-enhanced mutual authentication with session tokens, HITL approval
- `p2p-firewall`: Knowledge firewall with ACL rules, response sanitization, ZK attestation
- `p2p-protocol`: A2A message exchange over libp2p streams, remote agent adapter
- `p2p-discovery`: GossipSub agent card propagation, DHT agent advertisements, capability search
- `zkp-core`: gnark-based ProverService with PlonK/Groth16, ownership/balance/attestation/capability circuits
- `p2p-payment`: Peer-to-peer USDC payment with session verification

### Modified Capabilities
- `a2a-protocol`: Agent Card extended with DID, multiaddrs, capabilities, pricing, ZK credentials
- `blockchain-wallet`: PublicKey() method added to WalletProvider interface

## Impact

- **New packages**: `internal/p2p/` (node, identity, handshake, firewall, protocol, discovery), `internal/zkp/` (circuits)
- **Modified files**: config/types.go, wallet/*.go, a2a/server.go, app/wiring.go, app/app.go, app/tools.go, app/types.go, orchestration/tools.go
- **Dependencies**: go-libp2p v0.47.0, go-libp2p-kad-dht v0.38.0, go-libp2p-pubsub v0.15.0, go-multiaddr v0.16.1, gnark v0.14.0
- **Config**: New `p2p` section with enabled flag, listen addresses, bootstrap peers, firewall rules, ZK settings
- **Orchestration**: vault agent now routes `p2p_*` tools
