# Security Commands

Commands for managing encryption, secrets, and security configuration. See the [Security](../security/index.md) section for detailed documentation.

```
lango security <subcommand>
```

---

## lango security status

Show the current security configuration status. By default, runs in **passphrase-free mode** — reads `envelope.json` directly and attempts a non-interactive DB read via keyring/keyfile. DB-dependent fields gracefully degrade to zero/"unavailable" when no credential is available.

```
lango security status [--json] [--full]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |
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
| `signer_provider` | string | Active signer provider (`local`, `unavailable`) |
| `encryption_keys` | int | Number of registered encryption keys |
| `stored_secrets` | int | Number of stored encrypted secrets |
| `interceptor` | string | Interceptor status (`enabled`/`disabled`) |
| `pii_redaction` | string | PII redaction status (`enabled`/`disabled`) |
| `approval_policy` | string | Tool approval policy (`always`, `dangerous`, `never`) |
| `db_encryption` | string | Database encryption status |
| `db_available` | bool | Whether DB was accessible non-interactively |
| `envelope` | object | Envelope details (see below) |
| `kms_provider` | string | KMS provider name (when configured) |
| `kms_key_id` | string | KMS key identifier (when configured) |
| `kms_fallback` | string | KMS fallback status (when configured) |

**Envelope JSON fields:**

| Field | Type | Description |
|-------|------|-------------|
| `present` | bool | Whether envelope.json exists |
| `version` | int | Envelope format version |
| `slot_count` | int | Number of KEK slots |
| `slot_types` | []string | Unique slot types (`passphrase`, `mnemonic`) |
| `recovery_setup` | bool | Whether a mnemonic recovery slot exists |
| `pending_migration` | bool | Data re-encryption incomplete |
| `pending_rekey` | bool | SQLCipher rekey incomplete |

---

## lango security change-passphrase

Change the passphrase by re-wrapping the Master Key. No data is re-encrypted and no DB rekey is issued — the operation is O(1) regardless of data size.

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

```
lango security migrate-passphrase
```

---

## Recovery Mnemonic

Manage the BIP39 recovery mnemonic for the Master Key envelope.

### lango security recovery setup

Generate a 24-word BIP39 recovery mnemonic and add it as a KEK slot. The mnemonic is displayed exactly once — you must write it down and store it securely.

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

Have you written down all 24 words? [y/N] y
Enter word 7 to confirm: absorb
Enter word 19 to confirm: ...
Recovery mnemonic slot added successfully.
```

---

### lango security recovery restore

Recover access using the BIP39 mnemonic when the passphrase is lost. Unwraps the Master Key via the mnemonic slot and sets a new passphrase.

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

Store the master passphrase using the best available secure hardware backend. Requires an interactive terminal and a hardware backend (Touch ID or TPM 2.0).

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

Remove the master passphrase from all hardware keyring backends.

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
Remove passphrase from all keyring backends? [y/N] y
Removed passphrase from secure provider.

# Non-interactive
$ lango security keyring clear --force
Removed passphrase from secure provider.
```

---

### lango security keyring status

Show hardware keyring availability and stored passphrase status.

```
lango security keyring status [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

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

## Database Encryption

Encrypt or decrypt the application database using SQLCipher.

### lango security db-migrate

Convert the plaintext SQLite database to SQLCipher-encrypted format using the current passphrase.

```
lango security db-migrate [--force]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | bool | `false` | Skip confirmation prompt (enables non-interactive mode) |

**Example:**

```bash
$ lango security db-migrate
This will encrypt your database. A backup will be created. Continue? [y/N] y
Enter passphrase for DB encryption: ********
Encrypting database...
Database encrypted successfully.
Set security.dbEncryption.enabled=true in your config to use the encrypted DB.
```

---

### lango security db-decrypt

Convert a SQLCipher-encrypted database back to plaintext SQLite.

```
lango security db-decrypt [--force]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | bool | `false` | Skip confirmation prompt (enables non-interactive mode) |

**Example:**

```bash
$ lango security db-decrypt
This will decrypt your database to plaintext. Continue? [y/N] y
Enter passphrase for DB decryption: ********
Decrypting database...
Database decrypted successfully.
Set security.dbEncryption.enabled=false in your config if you no longer want encryption.
```

---

## Cloud KMS / HSM

Manage Cloud KMS and HSM integration. Requires `security.signer.provider` to be set to a KMS provider (`aws-kms`, `gcp-kms`, `azure-kv`, or `pkcs11`).

### lango security kms status

Show the KMS provider connection status.

```
lango security kms status [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

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

Test KMS encrypt/decrypt roundtrip using 32 bytes of random data.

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

List KMS keys registered in the KeyRegistry.

```
lango security kms keys [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango security kms keys
ID                                    NAME                  TYPE          REMOTE KEY ID
550e8400-e29b-41d4-a716-446655440000  primary-signing       signing       arn:aws:kms:us-east-1:...
6ba7b810-9dad-11d1-80b4-00c04fd430c8  default-encryption    encryption    arn:aws:kms:us-east-1:...
```

---

## Secret Management

Manage encrypted secrets stored in the database. Secret values are never displayed -- only metadata is shown when listing.

### lango security secrets list

List all stored secrets. Values are never shown.

```
lango security secrets list [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

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

Store a new encrypted secret or update an existing one. In interactive mode, prompts for the secret value (input is hidden). In non-interactive mode, use `--value-hex` to provide a hex-encoded value.

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

Delete a stored secret. Prompts for confirmation unless `--force` is specified.

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
Delete secret 'my-api-key'? [y/N] y
Secret 'my-api-key' deleted.

# Non-interactive
$ lango security secrets delete my-api-key --force
Secret 'my-api-key' deleted.
```

!!! tip
    Use `--force` for non-interactive environments (scripts, CI/CD). Without it, the command fails in non-interactive terminals.
