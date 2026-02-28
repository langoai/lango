## MODIFIED Requirements

### Requirement: TPMProvider seals secrets with TPM 2.0

The TPM provider SHALL use `TPMTSymDefObject` for the SRK template symmetric parameters. The provider SHALL use `tpm2.Marshal` with single-return signature and `tpm2.Unmarshal` with generic type parameter signature `Unmarshal[T]([]byte) (*T, error)` as required by go-tpm v0.9.8.

#### Scenario: SRK template uses correct symmetric type
- **WHEN** the TPM provider creates a primary key with ECC P256 SRK template
- **THEN** the template's `Symmetric` field SHALL be of type `TPMTSymDefObject`

#### Scenario: Marshal sealed blob without error return
- **WHEN** the TPM provider marshals `TPM2BPublic` and `TPM2BPrivate` to bytes
- **THEN** the system SHALL call `tpm2.Marshal` which returns `[]byte` directly

#### Scenario: Unmarshal sealed blob with generic type parameter
- **WHEN** the TPM provider unmarshals bytes into `TPM2BPublic` or `TPM2BPrivate`
- **THEN** the system SHALL call `tpm2.Unmarshal[T](data)` returning `(*T, error)` and dereference the result
