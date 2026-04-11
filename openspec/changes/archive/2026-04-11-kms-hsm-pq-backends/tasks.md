# Tasks: KMS/HSM PQ Backends

## Wave 1 — Envelope KMS Slot (core crypto)

- [x] Add `WrapAlgKMSEnvelope = "kms-envelope"` constant to `envelope.go`
- [x] Add `KMSProvider`, `KMSKeyID` fields to `KEKSlot` struct (omitempty)
- [x] Update `domainForSlotType` for `KEKSlotHardware` → `"kms"`
- [x] Implement `AddKMSSlot(ctx, label, mk, provider, kmsProviderName, kmsKeyID)` on `MasterKeyEnvelope`
- [x] Implement `UnwrapFromKMS(ctx, provider, providerName, keyID)` with 2-tier matching
- [x] Add `ErrKMSSlotUnavailable` sentinel to `errors.go`
- [x] Add `MasterKey()` accessor to `LocalCryptoProvider`
- [x] Test: `TestEnvelope_AddKMSSlot` — add KMS slot, verify fields
- [x] Test: `TestEnvelope_UnwrapFromKMS_RoundTrip` — add + unwrap, verify MK matches
- [x] Test: `TestEnvelope_UnwrapFromKMS_Failure` — mock decrypt error → `ErrUnwrapFailed`
- [x] Test: `TestEnvelope_UnwrapFromKMS_TierMatching` — Tier 1 exact, Tier 2 provider-only, no match
- [x] Test: `TestEnvelope_KMSSlot_JSON_RoundTrip` — serialize/deserialize KMS fields
- [x] Test: `TestEnvelope_KMSSlot_BackwardCompat` — old envelope without KMS fields loads OK
- [x] Verify: `go build ./... && go test ./internal/security/...`

## Wave 2 — Bootstrap integration

- [x] Create `internal/bootstrap/kms_env.go` with `KMSConfigFromEnv()` (provider-specific env vars)
- [x] Add `KMSProvider`, `KMSUnwrap` fields to bootstrap `State`
- [x] Add `KMSConfig`, `KMSProviderName` fields to bootstrap `Options`
- [x] Add `KMSUnwrap` field to bootstrap `Result`
- [x] Update `Run()`: explicit Options > env vars (KMSConfigFromEnv fallback)
- [x] Update `phaseAcquireCredential`: KMS detection + bare KMS unwrap before passphrase
- [x] Test: KMS bootstrap path — mock KMS + envelope with KMS slot → MK unwrapped
- [x] Test: KMS fallback — no KMS slot → passphrase path
- [x] Test: `KMSConfigFromEnv` — provider-specific env parsing (aws, azure, pkcs11, none)
- [x] Verify: `go build ./... && go test ./internal/bootstrap/...`

## Wave 3 — CLI + Status

- [x] Add `newKMSWrapCmd` — `lango security kms wrap --provider <name> --key-id <id>`
- [x] Add `newKMSDetachCmd` — `lango security kms detach [--slot-id <uuid>]`
- [x] Add `KMSProtected`, `KMSProvider` to `envelopeSection` in `status.go`
- [x] Update text and JSON output for KMS protection status
- [x] Verify: `go build ./... && go test ./internal/cli/security/...`

## Final Verification

- [x] Full build: `go build ./... && go vet ./...`
- [x] Full test: `go test ./internal/security/... ./internal/bootstrap/... ./internal/cli/security/... -count=1`
