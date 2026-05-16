# CLI ZKP Inspection

## Purpose
Provides CLI commands for inspecting Zero-Knowledge Proof configuration and available circuits.

## Requirements

### Requirement: ZKP status command
The system SHALL provide a `lango p2p zkp status [--output table|json]` command that displays the current ZKP configuration including handshake and attestation state, SRS mode, proving scheme, proof cache location, and maximum credential age. The command SHALL use cfgLoader (config only) since it reads configuration state.

#### Scenario: ZKP enabled
- **WHEN** user runs `lango p2p zkp status` with ZKP enabled
- **THEN** system displays zkp.enabled, srsMode, srsPath, maxCredentialAge, and proof scheme

#### Scenario: ZKP status with JSON output
- **WHEN** user runs `lango p2p zkp status --output json`
- **THEN** system outputs a JSON object with fields: zkHandshake, zkAttestation, provingScheme, srsMode, srsPath, proofCacheDir, maxCredentialAge

### Requirement: ZKP circuits command
The system SHALL provide a `lango p2p zkp circuits [--output table|json]` command that lists all available ZK circuits with their names and descriptions. The command SHALL NOT require any bootLoader or cfgLoader since it returns static data compiled into the binary.

#### Scenario: List circuits in text format
- **WHEN** user runs `lango p2p zkp circuits`
- **THEN** system displays a table with CIRCUIT and DESCRIPTION columns listing all registered circuits (identity, capability, attestation, reputation)

#### Scenario: List circuits in JSON format
- **WHEN** user runs `lango p2p zkp circuits --output json`
- **THEN** system outputs a JSON array of circuit objects with name and description fields

### Requirement: ZKP command group entry
The system SHALL provide a `lango p2p zkp` command group that shows help text listing all ZKP subcommands when invoked without a subcommand.

#### Scenario: Help text
- **WHEN** user runs `lango p2p zkp`
- **THEN** system displays help listing status and circuits subcommands
