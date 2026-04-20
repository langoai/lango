---
title: Security
---

# Security

Lango provides multiple layers of security to protect sensitive data flowing between users, agents, and AI providers.

## Security Layers

| Layer | Purpose | Details |
|-------|---------|---------|
| **Master Key Envelope** | Hierarchical key management | MK/KEK architecture, recovery mnemonic, passphrase rotation without re-encryption |
| **Encryption & Secrets** | Protect data at rest and in transit | AES-256-GCM encryption, key registry, secret management |
| **PII Redaction** | Strip personal information before it reaches AI providers | Regex patterns + optional NER via Microsoft Presidio |
| **Exportability Policy** | Decide what early knowledge artifacts can be traded | Source-primary evaluation for knowledge exchange v1 |
| **Approval Flow** | Decide when early artifact releases can move forward | Structured artifact release states and audit-backed receipts |
| **Tool Approval** | Control which tools agents can execute | Policy-based approval workflows with channel notifications |
| **Authentication** | Secure gateway access | OIDC login flow, session management, CORS controls |
| **Hardware Keyring** | Secure passphrase storage | Hardware-backed passphrase via Touch ID (macOS Secure Enclave) or TPM 2.0 (Linux) |
| **Payload Protection** | Protect sensitive data at rest | Broker-managed AES-256-GCM payload encryption with redacted search projections |
| **Cloud KMS / HSM** | Hardware-backed cryptography | AWS KMS, GCP KMS, Azure Key Vault, PKCS#11 HSM integration |
| **P2P Session Management** | Peer session lifecycle | Session listing, explicit invalidation, security-event-based revocation |
| **P2P Tool Sandbox** | Execution isolation | Subprocess and container-based isolation for remote tool invocations |
| **P2P Auth Hardening** | Signed challenge protocol | ECDSA signed challenges, nonce replay protection, timestamp validation |

## Architecture

```mermaid
graph LR
    User -->|input| Interceptor
    Interceptor -->|PII redacted| Agent
    Agent -->|tool call| Approval[Tool Approval]
    Approval -->|approved| Tool
    Tool -->|secret ref| RefStore[Secret RefStore]
    RefStore -->|resolved| Execution
    Agent -->|output| Scanner[Output Scanner]
    Scanner -->|secrets masked| User
```

The **security interceptor** sits between the user and the AI agent. It:

1. **Redacts PII** from user input before forwarding to the AI provider
2. **Gates tool execution** behind an approval workflow for sensitive operations
3. **Scans agent output** to replace any leaked secret values with `[SECRET:name]` placeholders

!!! info "Enable the Interceptor"

    The security interceptor is disabled by default. Enable it in your configuration:

    > **Settings:** `lango settings` → Security

    ```json
    {
      "security": {
        "interceptor": {
          "enabled": true,
          "redactPii": true,
          "approvalPolicy": "dangerous"
        }
      }
    }
    ```

## Encryption Modes

Lango supports three encryption modes depending on your deployment:

- **Local Mode** (default) -- AES-256-GCM with a Master Key envelope. The passphrase wraps a random Master Key; data keys are derived from the MK. Supports recovery mnemonic and O(1) passphrase rotation.
- **RPC Mode** (production) -- Delegates cryptographic operations to a hardware-backed companion app or external signer. Keys never leave secure hardware.
- **Cloud KMS Mode** (enterprise) -- Delegates to managed key services (AWS KMS, GCP KMS, Azure Key Vault) or on-premises HSM via PKCS#11. Automatic fallback to local mode when KMS is unavailable.

See [Encryption & Secrets](encryption.md) for full details.

## Quick Links

- [Master Key Envelope](envelope-migration.md) -- MK/KEK hierarchy, migration from legacy, backup & recovery
- [Encryption & Secrets](encryption.md) -- Key derivation, secret storage, output scanning, companion app
- [PII Redaction](pii-redaction.md) -- Builtin patterns, custom regex, Presidio integration
- [Exportability Policy](exportability.md) -- Source-primary exportability evaluation for early knowledge exchange
- [Approval Flow](approval-flow.md) -- Structured artifact release decisions and audit-backed receipts
- [Dispute-Ready Receipts](dispute-ready-receipts.md) -- Lite submission and transaction receipts for early knowledge exchange
- [Tool Approval](tool-approval.md) -- Approval policies, sensitive/exempt tools, notifications
- [Authentication](authentication.md) -- OIDC providers, session management, CORS configuration
- [Hardware Keyring](encryption.md#hardware-keyring-integration) -- Secure passphrase storage via Touch ID / TPM
- [Payload Protection](encryption.md#database-encryption) -- Broker-managed payload encryption and legacy DB compatibility notes
- [Cloud KMS / HSM](encryption.md#cloud-kms-mode) -- AWS, GCP, Azure, PKCS#11 integration
