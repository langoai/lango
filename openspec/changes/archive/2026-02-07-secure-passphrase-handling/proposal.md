## Why

Currently, `LocalCryptoProvider` reads the passphrase from the config file. If an AI Agent reads the config via filesystem tools, the passphrase can be exposed. This is inconsistent with the security goal that "AI must not have access to encryption keys".

## What Changes

- **Interactive Passphrase Prompt**: Accept input directly from terminal instead of reading passphrase from config
- **Passphrase Checksum**: Store checksum with salt for early detection of incorrect passphrase input
- **Migration Process**: Process for re-encrypting existing encrypted data with a new key when passphrase is changed
- **Security Mode Documentation**: Clearly document LocalCryptoProvider (dev/test) and RPCProvider+Companion (production) modes
- **Doctor Warnings**: Show development/test-only warnings when using local provider

## Capabilities

### New Capabilities
- `passphrase-management`: Interactive passphrase prompt, checksum validation, migration workflow

### Modified Capabilities
- `secure-signer`: LocalCryptoProvider initialization logic changed (config → interactive prompt)
- `cli-doctor`: Security provider mode warnings/recommendations added

## Impact

- `internal/app/app.go`: LocalCryptoProvider initialization logic modified
- `internal/config/types.go`: Passphrase field deprecated
- `internal/session/ent_store.go`: Checksum storage/verification methods added
- `internal/security/local_provider.go`: Migration logic added
- `internal/cli/doctor/checks/security.go`: Provider mode check added
- `README.md`: Two mode descriptions added to Security section
