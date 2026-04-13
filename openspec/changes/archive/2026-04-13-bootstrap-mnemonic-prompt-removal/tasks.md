## Tasks

### 1. Decouple `recovery restore` from bootstrap

- [x] 1.1 Modify `newRecoveryRestoreCmd()` in `internal/cli/security/recovery.go` to remove `bootLoader` parameter
- [x] 1.2 Replace `bootLoader()` call with `defaultLangoDir()` + `security.LoadEnvelopeFile(langoDir)` direct loading
- [x] 1.3 Add `langoDir` empty-string check and `envelope == nil` check with error `"envelope not found — recovery requires local encryption mode"`
- [x] 1.4 Remove `boot.Crypto.(*security.LocalCryptoProvider)` type assertion and `defer boot.DBClient.Close()`
- [x] 1.5 Replace all `boot.LangoDir` references with local `langoDir` variable
- [x] 1.6 Update `newRecoveryCmd(bootLoader)` to call `newRecoveryRestoreCmd()` without arguments while keeping `newRecoverySetupCmd(bootLoader)` unchanged

### 2. Remove mnemonic prompt from bootstrap

- [x] 2.1 Delete mnemonic recovery block in `internal/bootstrap/phases.go` lines 168-191 (comment + if block)
- [x] 2.2 Update `phaseAcquireCredential` doc comment (lines 125-126) to: `"attempts KMS unwrap first (if configured), then falls back to the passphrase acquisition chain"`
- [x] 2.3 Update `phaseUnwrapOrCreateMK` doc comment (lines 228-231): case 1 from `"mnemonic recovery"` to `"KMS unwrap"`
- [x] 2.4 Update line 239 comment from `"Already unwrapped via mnemonic recovery."` to `"Already unwrapped (e.g. KMS)."`
- [x] 2.5 Update `RecoveryMode` comment in `internal/bootstrap/pipeline.go` line 55 to `"reserved for future use; no longer set during bootstrap"`

### 3. Tests

- [x] 3.1 Add test in `internal/cli/security/recovery_test.go`: restore with temp dir envelope loads envelope directly without bootstrap
- [x] 3.2 Add test in `internal/cli/security/recovery_test.go`: restore with no envelope returns `"envelope not found"` error
- [x] 3.3 Add behavioral test in `internal/bootstrap/bootstrap_envelope_test.go`: bootstrap with mnemonic-slot envelope proceeds via passphrase path normally

### 4. Verification

- [x] 4.1 `go build ./...` passes
- [x] 4.2 `go test ./...` passes
