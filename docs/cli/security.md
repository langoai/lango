# Security Commands

Commands for managing encryption, secrets, and security configuration. See the [Security](../security/index.md) section for detailed documentation.

```
lango security <subcommand>
```

---

## lango security status

Show the current security configuration status. By default, runs in **passphrase-free mode** — reads `envelope.json` directly and attempts a non-interactive DB read via keyring/keyfile. DB-dependent fields gracefully degrade to zero/"unavailable" when no credential is available. The command writes through the Cobra command output stream so wrappers and test harnesses can capture table or JSON output directly.

```
lango security status [--output table|json] [--full]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | `table` | Output format (`table` or `json`) |
| `--full` | bool | `false` | Run full bootstrap (may prompt for passphrase) |

**Example:**

```bash
$ lango security status
Security Status
  Signer Provider:    local
  Encryption Keys:    2
  Stored Secrets:     5
  Interceptor:        enabled
  PII Redaction:      disabled
  Approval Policy:    dangerous
  DB Encryption:      disabled (plaintext)
  Master Key Envelope:
    Version:          1
    KEK Slots:        2 (passphrase, mnemonic)
    Recovery Setup:   enabled
```

```bash
# DB unavailable (no keyring/keyfile)
$ lango security status
Security Status
  Signer Provider:    unavailable
  Encryption Keys:    0
  Stored Secrets:     0
  ...
  DB Access:          unavailable (no non-interactive credential)
  Master Key Envelope:
    Version:          1
    KEK Slots:        1 (passphrase)
    Recovery Setup:   disabled
```

**JSON output fields:**

| Field | Type | Description |
|-------|------|-------------|
| `signer_provider` | string | Active signer provider (`local`, `rpc`, `aws-kms`, `gcp-kms`, `azure-kv`, `pkcs11`) or `unavailable` when DB-backed config could not be read non-interactively |
| `encryption_keys` | int | Number of registered encryption keys |
| `stored_secrets` | int | Number of stored encrypted secrets |
| `interceptor` | string | Interceptor status (`enabled`/`disabled`) |
| `pii_redaction` | string | PII redaction status (`enabled`/`disabled`) |
| `approval_policy` | string | Tool approval policy (`always`, `dangerous`, `never`) |
| `db_encryption` | string | Current DB-encryption / legacy-DB status (`disabled (plaintext)`, `legacy encrypted or unreadable DB (unsupported)`, or `deprecated config (ignored)`) |
| `db_available` | bool | Whether DB was accessible non-interactively |
| `envelope` | object | Envelope details (see below) |
| `kms_provider` | string | KMS provider name (when configured) |
| `kms_key_id` | string | KMS key identifier (when configured) |
| `kms_fallback` | string | KMS fallback status (`enabled`/`disabled`) when a KMS-backed signer is active |

**Envelope JSON fields:**

| Field | Type | Description |
|-------|------|-------------|
| `present` | bool | Whether envelope.json exists |
| `version` | int | Envelope format version |
| `slot_count` | int | Number of KEK slots |
| `slot_types` | []string | Unique slot types (`passphrase`, `mnemonic`) |
| `recovery_setup` | bool | Whether a mnemonic recovery slot exists |
| `pending_migration` | bool | Data re-encryption incomplete |
| `pending_rekey` | bool | Legacy pre-upgrade SQLCipher migration state |

---

## lango security change-passphrase

Change the passphrase by re-wrapping the Master Key. No data is re-encrypted and no DB rekey is issued — the operation is O(1) regardless of data size. The success confirmation is written through the Cobra command output stream, and any keyfile/keyring update notices or warnings are written through the Cobra command error stream so wrappers and test harnesses can capture them directly.

```
lango security change-passphrase
```

!!! info "Requirements"
    - Only available for envelope-based installations (local crypto provider)
    - Requires an interactive terminal
    - Recovery mnemonic slots are unchanged

**Process:**

1. Full bootstrap verifies the current passphrase
2. You enter and confirm the current passphrase again
3. You enter and confirm a new passphrase (min 8 characters)
4. The Master Key is re-wrapped with the new passphrase-derived KEK
5. The updated envelope is atomically persisted

**Example:**

```bash
$ lango security change-passphrase
Enter CURRENT passphrase: ********
Enter NEW passphrase: ********
Confirm NEW passphrase: ********
Passphrase changed. No data was re-encrypted.
```

---

## lango security migrate-passphrase (deprecated)

!!! warning "Deprecated"
    Use `lango security change-passphrase` instead. The legacy command re-encrypts all data, which is unnecessary with the envelope architecture.

Its non-error guidance, progress, and completion output are written through the Cobra command output stream so wrappers and test harnesses can capture migration progress directly.

```
lango security migrate-passphrase
```

---

## Recovery Mnemonic

Manage the BIP39 recovery mnemonic for the Master Key envelope.

### lango security recovery setup

Generate a 24-word BIP39 recovery mnemonic and add it as a KEK slot. The mnemonic is displayed exactly once — you must write it down and store it securely. The mnemonic banner, written-down confirmation prompt, confirmation-word prompt, and success message are written through the Cobra command output stream so wrappers and test harnesses can capture them directly, and the confirmation-word input itself goes through the shared prompt helper on top of the Cobra command streams.

```
lango security recovery setup
```

!!! warning "Requirements"
    - Requires an interactive terminal
    - Only one mnemonic slot is allowed per envelope
    - The current passphrase must be provided to authorize setup

**Process:**

1. Enter the current passphrase to unwrap the Master Key
2. A 24-word BIP39 mnemonic is generated and displayed
3. Confirm you have written it down
4. Enter two randomly selected words to verify
5. The mnemonic KEK slot is added to the envelope

**Example:**

```bash
$ lango security recovery setup
Enter current passphrase to authorize setup: ********

============================================================
RECOVERY MNEMONIC — write this down and store securely
============================================================
 1. abandon    2. ability    3. able       4. about
 5. above      6. absent     7. absorb     8. abstract
 ...
============================================================

Have you written down all 24 words? [y/N]: y
Enter word 7 to confirm: absorb
Enter word 19 to confirm: ...
Recovery mnemonic slot added successfully.
```

The confirmation-word prompts also accept the final matching line when a wrapper-driven input stream ends without a trailing newline, so seam-driven execution stays aligned with interactive success behavior.

---

### lango security recovery restore

Recover access using the BIP39 mnemonic when the passphrase is lost. Unwraps the Master Key via the mnemonic slot and sets a new passphrase. The success confirmation is written through the Cobra command output stream, and any keyfile/keyring update notices or warnings are written through the Cobra command error stream so wrappers and test harnesses can capture them directly.

```
lango security recovery restore
```

!!! info "Requirements"
    - Requires an interactive terminal
    - A mnemonic slot must exist on the envelope

**Process:**

1. Enter the 24-word recovery mnemonic
2. The mnemonic is validated and used to unwrap the Master Key
3. Enter and confirm a new passphrase
4. The passphrase KEK slot is replaced with the new passphrase
5. The recovery mnemonic slot is unchanged

**Example:**

```bash
$ lango security recovery restore
Enter 24-word recovery mnemonic: ********
Enter NEW passphrase: ********
Confirm NEW passphrase: ********
Recovery complete. The new passphrase is now active.
```

---

## Hardware Keyring

Manage hardware-backed keyring passphrase storage. Only secure hardware backends are supported (macOS Touch ID / Linux TPM 2.0) to prevent same-UID attacks.

### lango security keyring store

Store the master passphrase using the best available secure hardware backend. Requires an interactive terminal and a hardware backend (Touch ID or TPM 2.0). Non-error status messages are written through the Cobra command output stream so wrappers and test harnesses can capture them directly.

```
lango security keyring store
```

!!! warning "Requirements"
    - An interactive terminal (cannot be used in CI/CD)
    - A secure hardware backend (Touch ID on macOS or TPM 2.0 on Linux)
    - On macOS: binary must be codesigned for biometric protection

**Example:**

```bash
$ lango security keyring store
Enter passphrase to store: ********
Passphrase stored with biometric protection.
  Next launch will load it automatically.
```

---

### lango security keyring clear

Remove the master passphrase from all hardware keyring backends. The command uses the shared confirmation helper on top of Cobra command streams for prompt output, confirmation input, and result messaging, so wrappers and test harnesses can drive the interaction through `cmd.OutOrStdout()`, `cmd.InOrStdin()`, and `cmd.ErrOrStderr()`. In non-interactive runs, pass `--force` instead of attempting a prompt.

```
lango security keyring clear [--force]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | bool | `false` | Skip confirmation prompt |

**Examples:**

```bash
# Interactive
$ lango security keyring clear
Remove passphrase from all keyring backends? [y/N]: y
Removed passphrase from secure provider.

# Non-interactive
$ lango security keyring clear --force
Removed passphrase from secure provider.
```

---

### lango security keyring status

Show hardware keyring availability and stored passphrase status.

```
lango security keyring status [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango security keyring status
Hardware Keyring Status
  Available:       true
  Security Tier:   biometric
  Has Passphrase:  true
```

**JSON output fields:**

| Field | Type | Description |
|-------|------|-------------|
| `available` | bool | Whether a hardware keyring is available |
| `security_tier` | string | Security tier (`biometric`, `tpm`, or `none`) |
| `has_passphrase` | bool | Whether passphrase is stored |

---

## Legacy Database Encryption

Legacy SQLCipher database workflows are no longer supported by the current runtime. The commands remain only as remediation signposts for users who need to export old databases with an older build first.

### lango security db-migrate

Legacy SQLCipher migration command. The current runtime exits with an unsupported/remediation message.

```
lango security db-migrate [--force]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | bool | `false` | Skip confirmation prompt (enables non-interactive mode) |

**Example:**

```bash
$ lango security db-migrate
SQLCipher database encryption is no longer supported by this runtime.
Use an older build to export or decrypt the legacy database before upgrading.
```

---

### lango security db-decrypt

Legacy SQLCipher decrypt command. The current runtime exits with an unsupported/remediation message.

```
lango security db-decrypt [--force]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | bool | `false` | Skip confirmation prompt (enables non-interactive mode) |

**Example:**

```bash
$ lango security db-decrypt
SQLCipher database decryption is no longer supported by this runtime.
Use an older build to export the legacy database before upgrading.
```

---

## Cloud KMS / HSM

Manage Cloud KMS and HSM integration. Requires `security.signer.provider` to be set to a KMS provider (`aws-kms`, `gcp-kms`, `azure-kv`, or `pkcs11`). The running binary must also include the matching KMS build tag, and the runtime still depends on bootstrap-backed storage wiring for the key registry and secrets store.

### lango security kms status

Show the KMS provider connection status. The command writes through the Cobra command output stream so wrappers and test harnesses can capture table or JSON output directly.

```
lango security kms status [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango security kms status
KMS Status
  Provider:      aws-kms
  Key ID:        arn:aws:kms:us-east-1:123456789012:key/example-key
  Region:        us-east-1
  Fallback:      enabled
  Status:        connected
```

**JSON output fields:**

| Field | Type | Description |
|-------|------|-------------|
| `provider` | string | KMS provider name |
| `key_id` | string | KMS key identifier |
| `region` | string | Cloud region (if applicable) |
| `fallback` | string | Local fallback status (`enabled`/`disabled`) |
| `status` | string | Connection status (`connected`, `unreachable`, `not configured`, or error) |

---

### lango security kms test

Test KMS encrypt/decrypt roundtrip using 32 bytes of random data. The command writes its progress and success output through the Cobra command output stream so wrappers and test harnesses can capture it directly.

```
lango security kms test
```

**Example:**

```bash
$ lango security kms test
Testing KMS roundtrip with key "arn:aws:kms:us-east-1:123456789012:key/example-key"...
  Encrypt: OK (32 bytes → 64 bytes)
  Decrypt: OK (32 bytes)
  Roundtrip: PASS
```

---

### lango security kms keys

List KMS keys registered in the KeyRegistry. The command writes through the Cobra command output stream so wrappers and test harnesses can capture empty-state, table, or JSON output directly.

```
lango security kms keys [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango security kms keys
ID                                    NAME                  TYPE          REMOTE KEY ID
550e8400-e29b-41d4-a716-446655440000  primary-signing       signing       arn:aws:kms:us-east-1:...
6ba7b810-9dad-11d1-80b4-00c04fd430c8  default-encryption    encryption    arn:aws:kms:us-east-1:...
```

---

### lango security kms wrap

Add a KMS KEK slot to protect the Master Key. Success output is written through the Cobra command output stream so wrappers and test harnesses can capture it directly.

```
lango security kms wrap --provider <name> --key-id <id>
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--provider` | string | *required* | KMS provider name (`aws-kms`, `gcp-kms`, `azure-kv`, `pkcs11`) |
| `--key-id` | string | *required* | KMS key identifier |

**Example:**

```bash
$ lango security kms wrap --provider aws-kms --key-id arn:aws:kms:us-east-1:123456789012:key/example
KMS slot added (provider=aws-kms, keyID=arn:aws:kms:us-east-1:123456789012:key/example)
Next bootstrap can use KMS for passphraseless unlock.
```

---

### lango security kms detach

Remove a KMS KEK slot from the envelope. Success output and multi-slot guidance are written through the Cobra command output stream so wrappers and test harnesses can capture them directly.

```
lango security kms detach [--slot-id <uuid>]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--slot-id` | string | `""` | UUID of the KMS slot to remove when multiple KMS slots exist |

**Examples:**

```bash
$ lango security kms detach
KMS slot 550e8400-e29b-41d4-a716-446655440000 removed.

$ lango security kms detach
Multiple KMS slots found. Specify --slot-id:
  550e8400-e29b-41d4-a716-446655440000  provider=aws-kms  keyID=key-a  label=kms-a
  6ba7b810-9dad-11d1-80b4-00c04fd430c8  provider=aws-kms  keyID=key-b  label=kms-b
```

---

## Secret Management

Manage encrypted secrets stored in the database. Secret values are never displayed -- only metadata is shown when listing.

### lango security secrets list

List all stored secrets. Values are never shown. The command writes through the Cobra command output stream so wrappers and test harnesses can capture table or JSON output directly.

```
lango security secrets list [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango security secrets list
NAME               KEY      CREATED           UPDATED           ACCESS_COUNT
anthropic-api-key  default  2026-01-15 10:00  2026-02-20 14:30  42
telegram-token     default  2026-01-15 10:05  2026-01-15 10:05  15
openai-api-key     default  2026-02-01 09:00  2026-02-01 09:00  3
```

---

### lango security secrets set

Store a new encrypted secret or update an existing one. In interactive mode, prompts for the secret value (input is hidden). In non-interactive mode, use `--value-hex` to provide a hex-encoded value. Success output is written through the Cobra command output stream.

```
lango security secrets set <name> [--value-hex <hex>]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Name identifier for the secret |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--value-hex` | string | - | Hex-encoded value to store (optional `0x` prefix). Enables non-interactive mode. |

**Examples:**

```bash
# Interactive (prompts for value)
$ lango security secrets set my-api-key
Enter secret value:
Secret 'my-api-key' stored successfully.

# Non-interactive with hex value (e.g., wallet private key in Docker/CI)
$ lango security secrets set wallet.privatekey --value-hex 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
Secret 'wallet.privatekey' stored successfully.

# Without 0x prefix
$ lango security secrets set wallet.privatekey --value-hex ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
Secret 'wallet.privatekey' stored successfully.
```

!!! tip
    Use `--value-hex` for non-interactive environments (Docker, CI/CD, scripts). Without it, the command requires an interactive terminal and will fail with an error suggesting `--value-hex`.

---

### lango security secrets delete

Delete a stored secret. Prompts for confirmation unless `--force` is specified. The command uses the shared confirmation helper on top of Cobra command streams for prompt output, confirmation input, and result messaging, so wrappers and test harnesses can drive the interaction through `cmd.OutOrStdout()`, `cmd.InOrStdin()`, and `cmd.ErrOrStderr()`. In non-interactive runs, pass `--force` instead of attempting a prompt.

```
lango security secrets delete <name> [--force]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Name of the secret to delete |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | bool | `false` | Skip confirmation prompt |

**Examples:**

```bash
# Interactive confirmation
$ lango security secrets delete my-api-key
Delete secret 'my-api-key'? [y/N]: y
Secret 'my-api-key' deleted.

# Non-interactive
$ lango security secrets delete my-api-key --force
Secret 'my-api-key' deleted.
```

!!! tip
    Use `--force` for non-interactive environments (scripts, CI/CD). Without it, the command fails in non-interactive terminals.
