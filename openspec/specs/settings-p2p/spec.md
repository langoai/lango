## Purpose

Define the TUI settings forms for P2P networking configuration: core network settings, ZKP, pricing, owner protection, and tool isolation sandbox.

## Requirements

### Requirement: P2P Network settings form
The settings TUI SHALL provide a "P2P Network" menu category with a form exposing core P2P configuration fields: enabled, listen addresses, bootstrap peers, relay, mDNS, max peers, handshake timeout, session token TTL, auto-approve known peers, gossip interval, ZK handshake, ZK attestation, require signed challenge, and min trust score.

#### Scenario: User enables P2P networking
- **WHEN** user navigates to "P2P Network" and sets Enabled to true
- **THEN** the config's `p2p.enabled` field SHALL be set to true upon save

#### Scenario: User sets listen addresses
- **WHEN** user enters comma-separated multiaddrs in "Listen Addresses"
- **THEN** the config's `p2p.listenAddrs` SHALL contain each address as a separate array element

### Requirement: P2P ZKP settings form
The settings TUI SHALL provide a "P2P ZKP" menu category with fields for proof cache directory, proving scheme (plonk/groth16), SRS mode (unsafe/file), SRS path, and max credential age.

#### Scenario: User selects groth16 proving scheme
- **WHEN** user selects "groth16" from the proving scheme dropdown
- **THEN** the config's `p2p.zkp.provingScheme` SHALL be set to "groth16"

### Requirement: P2P Pricing settings form
The settings TUI SHALL provide a "P2P Pricing" menu category with fields for enabled, price per query, and tool-specific prices (as key:value comma-separated text).

#### Scenario: User sets tool prices
- **WHEN** user enters "exec:0.10,browser:0.50" in the Tool Prices field
- **THEN** the config's `p2p.pricing.toolPrices` SHALL be a map with keys "exec" and "browser" and respective values

### Requirement: P2P Owner Protection settings form
The settings TUI SHALL provide a "P2P Owner Protection" menu category with fields for owner name, email, phone, extra terms, and block conversations.

#### Scenario: User sets block conversations with nil default
- **WHEN** the config's `blockConversations` is nil
- **THEN** the form SHALL display the checkbox as checked (default true)

#### Scenario: User unchecks block conversations
- **WHEN** user unchecks "Block Conversations"
- **THEN** the config's `p2p.ownerProtection.blockConversations` SHALL be a pointer to false

### Requirement: P2P Sandbox settings form
The settings TUI SHALL provide a "P2P Sandbox" menu category with fields for tool isolation (enabled, timeout, max memory) and container sandbox (enabled, runtime, image, network mode, read-only rootfs, CPU quota, pool size, pool idle timeout).

#### Scenario: User configures container sandbox
- **WHEN** user enables container sandbox and selects "docker" runtime
- **THEN** the config's `p2p.toolIsolation.container.enabled` SHALL be true and `runtime` SHALL be "docker"

#### Scenario: Container read-only rootfs defaults to true
- **WHEN** the config's `readOnlyRootfs` is nil
- **THEN** the form SHALL display the checkbox as checked (default true)
