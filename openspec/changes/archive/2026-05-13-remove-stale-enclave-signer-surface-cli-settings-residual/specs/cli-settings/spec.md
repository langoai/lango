## MODIFIED Requirements

### Requirement: Security form signer provider options
The Security form's signer provider dropdown SHALL include options for all supported providers: local, rpc, aws-kms, gcp-kms, azure-kv, pkcs11.

#### Scenario: Removed enclave provider is absent from signer dropdown
- **WHEN** user opens the Security form
- **THEN** the signer provider dropdown SHALL NOT include "enclave"
