## 1. Fix Type Compatibility

- [x] 1.1 Change `TPMTSymDef` to `TPMTSymDefObject` in srkTemplate() symmetric field (line 125)

## 2. Fix Marshal API

- [x] 2.1 Change `tpm2.Marshal(pub)` from two-return to single-return assignment in marshalSealedBlob
- [x] 2.2 Change `tpm2.Marshal(priv)` from two-return to single-return assignment in marshalSealedBlob

## 3. Fix Unmarshal API

- [x] 3.1 Change public unmarshal to generic `tpm2.Unmarshal[tpm2.TPM2BPublic](pubBytes)` with pointer dereference
- [x] 3.2 Change private unmarshal to generic `tpm2.Unmarshal[tpm2.TPM2BPrivate](privBytes)` with pointer dereference

## 4. Verification

- [x] 4.1 Run `GOOS=linux go build ./internal/keyring/...` and confirm zero errors
- [x] 4.2 Run `go build ./...` and confirm full project builds
