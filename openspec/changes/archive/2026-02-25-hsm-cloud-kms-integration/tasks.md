## 1. Foundation

- [x] 1.1 Add KMSConfig, AzureKVConfig, PKCS11Config structs to config/types.go
- [x] 1.2 Add KMS defaults to config/loader.go DefaultConfig and viper defaults
- [x] 1.3 Add KMS provider validation to config/loader.go Validate()
- [x] 1.4 Add KMS sentinel errors and KMSError type to security/errors.go
- [x] 1.5 Create kms_retry.go with withRetry exponential backoff
- [x] 1.6 Create kms_checker.go with KMSHealthChecker implementing ConnectionChecker
- [x] 1.7 Create kms_factory.go with NewKMSProvider dispatch function

## 2. Provider Implementations

- [x] 2.1 Create aws_kms_provider.go (build tag: kms_aws || kms_all)
- [x] 2.2 Create aws_kms_provider_stub.go (build tag: !kms_aws && !kms_all)
- [x] 2.3 Create gcp_kms_provider.go (build tag: kms_gcp || kms_all)
- [x] 2.4 Create gcp_kms_provider_stub.go (build tag: !kms_gcp && !kms_all)
- [x] 2.5 Create azure_kv_provider.go (build tag: kms_azure || kms_all)
- [x] 2.6 Create azure_kv_provider_stub.go (build tag: !kms_azure && !kms_all)
- [x] 2.7 Create pkcs11_provider.go (build tag: kms_pkcs11 || kms_all)
- [x] 2.8 Create pkcs11_provider_stub.go (build tag: !kms_pkcs11 && !kms_all)
- [x] 2.9 Create kms_all.go build tag grouping file

## 3. Tests

- [x] 3.1 Create kms_retry_test.go with transient/non-transient/exhaust/cancel tests
- [x] 3.2 Create aws_kms_provider_test.go (build tag: kms_aws || kms_all)
- [x] 3.3 Create gcp_kms_provider_test.go (build tag: kms_gcp || kms_all)
- [x] 3.4 Create azure_kv_provider_test.go (build tag: kms_azure || kms_all)
- [x] 3.5 Create pkcs11_provider_test.go (build tag: kms_pkcs11 || kms_all)

## 4. CLI and Wiring

- [x] 4.1 Create cli/security/kms.go with status, test, keys subcommands
- [x] 4.2 Register KMS command in NewSecurityCmd (migrate.go)
- [x] 4.3 Extend statusOutput in status.go with KMS fields
- [x] 4.4 Add KMS provider cases to initSecurity in wiring.go with CompositeCryptoProvider fallback

## 5. Documentation

- [x] 5.1 Update security-roadmap.md P2-9 to COMPLETED status
- [x] 5.2 Run OpenSpec workflow (ff, apply, verify, sync, archive)
