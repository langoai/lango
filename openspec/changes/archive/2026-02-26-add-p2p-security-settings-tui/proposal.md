## Why

The `lango settings` TUI editor exposes 21 configuration categories but is missing P2P networking, tool isolation/sandbox, keyring, DB encryption, and KMS settings. Users must manually edit encrypted config JSON to enable or tune these features, which is error-prone and inconsistent with the rest of the settings UX.

## What Changes

- Add 8 new menu categories to the settings TUI: P2P Network, P2P ZKP, P2P Pricing, P2P Owner Protection, P2P Sandbox, Security Keyring, Security DB Encryption, Security KMS
- Add form builders for each category with appropriate field types (bool, text, int, select, password)
- Add config write-back mappings for all ~53 new form fields
- Expand the existing Security form's signer provider options to include KMS backends (`aws-kms`, `gcp-kms`, `azure-kv`, `pkcs11`)
- Handle `*bool` config fields (ReadOnlyRootfs, BlockConversations) with `derefBool`/`boolPtr` helpers

## Capabilities

### New Capabilities
- `settings-p2p`: TUI forms for P2P network, ZKP, pricing, owner protection, and sandbox settings in the settings editor
- `settings-security-advanced`: TUI forms for keyring, DB encryption, and KMS settings in the settings editor

### Modified Capabilities
- `cli-settings`: Expanded menu with 8 new categories, updated signer provider options

## Impact

- `internal/cli/settings/menu.go` — 8 new menu entries
- `internal/cli/settings/forms_impl.go` — 8 new form builders + 2 helpers + signer option expansion
- `internal/cli/settings/editor.go` — 8 new case routes in handleMenuSelection()
- `internal/cli/tuicore/state_update.go` — ~53 new case entries + boolPtr helper
- `internal/cli/settings/forms_impl_test.go` — 13 new tests
