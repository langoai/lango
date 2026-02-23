## Purpose

Docker Compose integration example that proves 3 Lango agents can discover each other via P2P mDNS, establish DID identity, and transact USDC on a local Ethereum chain.

## Requirements

### Requirement: Docker Compose multi-agent environment
The `examples/p2p-trading/` directory SHALL contain a Docker Compose configuration that starts a local Ethereum node (Anvil), deploys a MockUSDC contract, and launches 3 Lango agents (Alice, Bob, Charlie) with P2P and payment enabled.

#### Scenario: All services start successfully
- **WHEN** `docker compose up -d` is run in the example directory
- **THEN** Anvil SHALL be healthy on port 8545, the setup service SHALL deploy MockUSDC and fund agents, and all 3 agents SHALL respond to `/health` within 90 seconds

### Requirement: MockUSDC contract
The `contracts/MockUSDC.sol` SHALL implement a minimal ERC-20 with `mint()`, `transfer()`, `transferFrom()`, `approve()`, `balanceOf()`, and `allowance()` functions with 6 decimals.

#### Scenario: Initial token distribution
- **WHEN** the setup script completes
- **THEN** each agent address SHALL have 1000 USDC (1000000000 smallest units)

### Requirement: P2P discovery between agents
The 3 agents SHALL discover each other via mDNS on the Docker bridge network within 15 seconds of startup.

#### Scenario: Peer discovery
- **WHEN** all agents have been running for 15 seconds
- **THEN** each agent's `GET /api/p2p/peers` SHALL report at least 2 connected peers

### Requirement: Extended Docker entrypoint
The `docker-entrypoint-p2p.sh` SHALL wait for the USDC contract address from the setup sidecar, substitute it into the config, import the config, and inject the wallet private key via `--value-hex` flag.

#### Scenario: Agent startup with key injection
- **WHEN** the agent container starts with AGENT_PRIVATE_KEY environment variable
- **THEN** the entrypoint SHALL store the private key via `lango security secrets set wallet.privatekey --value-hex` before starting the server

### Requirement: Integration test script
The `scripts/test-p2p-trading.sh` SHALL verify health, P2P status, peer discovery, DID identity, USDC balances, and a payment transfer via REST API and on-chain queries.

#### Scenario: End-to-end payment verification
- **WHEN** the test script executes a 1.00 USDC payment from Alice to Bob
- **THEN** Bob's on-chain USDC balance SHALL increase by 1000000 (1.00 USDC with 6 decimals)

### Requirement: Example configs include paid value exchange settings
The P2P trading example configs SHALL include pricing, owner protection, and minimum trust score settings.

#### Scenario: Alice config has pricing enabled
- **WHEN** user reads `examples/p2p-trading/configs/alice.json`
- **THEN** P2P section includes `pricing` object with enabled=true, perQuery="0.10", and toolPrices map

#### Scenario: Alice config has owner protection
- **WHEN** user reads `examples/p2p-trading/configs/alice.json`
- **THEN** P2P section includes `ownerProtection` object with ownerName="Alice" and blockConversations=true

#### Scenario: All configs have minTrustScore
- **WHEN** user reads any of alice.json, bob.json, or charlie.json
- **THEN** P2P section includes `minTrustScore: 0.3`

#### Scenario: Each agent has correct ownerName
- **WHEN** user reads bob.json
- **THEN** ownerProtection.ownerName is "Bob"
- **WHEN** user reads charlie.json
- **THEN** ownerProtection.ownerName is "Charlie"
