## MODIFIED Requirements

### Requirement: Settings menu categories
The settings TUI menu SHALL include all configuration categories. The menu SHALL contain categories for: Providers, Agent, Server, Channels, Tools, Session, Security, Auth, Knowledge, Skill, Observational Memory, Embedding & RAG, Graph Store, Multi-Agent, A2A Protocol, Payment, Cron Scheduler, Background Tasks, Workflow Engine, Librarian, P2P Network, P2P ZKP, P2P Pricing, P2P Owner Protection, P2P Sandbox, Security Keyring, Security DB Encryption, Security KMS, Save & Exit, and Cancel.

#### Scenario: Menu displays all 29 categories
- **WHEN** user opens the settings editor and navigates to the menu
- **THEN** the menu SHALL display 29 selectable categories plus Save & Exit and Cancel

#### Scenario: P2P categories appear after Librarian
- **WHEN** user scrolls through the menu
- **THEN** P2P and advanced security categories SHALL appear between Librarian and Save & Exit

### Requirement: Security form signer provider options
The Security form's signer provider dropdown SHALL include options for all supported providers: local, rpc, enclave, aws-kms, gcp-kms, azure-kv, pkcs11.

#### Scenario: KMS providers available in signer dropdown
- **WHEN** user opens the Security form
- **THEN** the signer provider dropdown SHALL include "aws-kms", "gcp-kms", "azure-kv", and "pkcs11" as options
