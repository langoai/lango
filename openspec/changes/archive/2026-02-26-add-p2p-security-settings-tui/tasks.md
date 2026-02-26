## 1. Menu & Routing

- [x] 1.1 Add 8 new categories to `NewMenuModel()` in `internal/cli/settings/menu.go`
- [x] 1.2 Add 8 new `case` entries in `handleMenuSelection()` in `internal/cli/settings/editor.go`

## 2. Form Builders

- [x] 2.1 Add `derefBool()` and `formatKeyValueMap()` helpers in `internal/cli/settings/forms_impl.go`
- [x] 2.2 Add `NewP2PForm()` — 14 fields for P2P Network settings
- [x] 2.3 Add `NewP2PZKPForm()` — 5 fields for ZKP settings
- [x] 2.4 Add `NewP2PPricingForm()` — 3 fields for pricing settings
- [x] 2.5 Add `NewP2POwnerProtectionForm()` — 5 fields for owner protection
- [x] 2.6 Add `NewP2PSandboxForm()` — 11 fields for tool isolation & container sandbox
- [x] 2.7 Add `NewKeyringForm()` — 1 field for OS keyring
- [x] 2.8 Add `NewDBEncryptionForm()` — 2 fields for SQLCipher encryption
- [x] 2.9 Add `NewKMSForm()` — 12 fields for Cloud KMS / HSM

## 3. Config Write-back

- [x] 3.1 Add `boolPtr()` helper in `internal/cli/tuicore/state_update.go`
- [x] 3.2 Add P2P Network case entries (~14) in `UpdateConfigFromForm()`
- [x] 3.3 Add P2P ZKP case entries (~5) in `UpdateConfigFromForm()`
- [x] 3.4 Add P2P Pricing case entries (~3) in `UpdateConfigFromForm()`
- [x] 3.5 Add P2P Owner Protection case entries (~5) in `UpdateConfigFromForm()`
- [x] 3.6 Add P2P Sandbox case entries (~11) in `UpdateConfigFromForm()`
- [x] 3.7 Add Security Keyring case entry in `UpdateConfigFromForm()`
- [x] 3.8 Add Security DB Encryption case entries (~2) in `UpdateConfigFromForm()`
- [x] 3.9 Add Security KMS case entries (~12) in `UpdateConfigFromForm()`

## 4. Existing Form Update

- [x] 4.1 Expand signer provider options in `NewSecurityForm()` to include KMS backends

## 5. Tests

- [x] 5.1 Add form field count/key tests for all 8 new forms
- [x] 5.2 Add menu category existence test for all 8 new categories
- [x] 5.3 Add config round-trip tests for P2P, Sandbox *bool, and KMS fields
- [x] 5.4 Add `derefBool` helper test

## 6. Verification

- [x] 6.1 Run `go build ./...` — no errors
- [x] 6.2 Run `go test ./internal/cli/settings/...` — all pass
