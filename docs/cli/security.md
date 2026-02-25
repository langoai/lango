# Security Commands

Commands for managing encryption, secrets, and security configuration. See the [Security](../security/index.md) section for detailed documentation.

```
lango security <subcommand>
```

---

## lango security status

Show the current security configuration status including signer provider, encryption keys, stored secrets count, and interceptor settings.

```
lango security status [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

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
```

```bash
# With KMS configured
$ lango security status
Security Status
  Signer Provider:    aws-kms
  Encryption Keys:    2
  Stored Secrets:     5
  Interceptor:        enabled
  PII Redaction:      disabled
  Approval Policy:    dangerous
  DB Encryption:      encrypted (active)
  KMS Provider:       aws-kms
  KMS Key ID:         arn:aws:kms:us-east-1:...
  KMS Fallback:       enabled
```

**JSON output fields:**

| Field | Type | Description |
|-------|------|-------------|
| `signer_provider` | string | Active signer provider (`local`) |
| `encryption_keys` | int | Number of registered encryption keys |
| `stored_secrets` | int | Number of stored encrypted secrets |
| `interceptor` | string | Interceptor status (`enabled`/`disabled`) |
| `pii_redaction` | string | PII redaction status (`enabled`/`disabled`) |
| `approval_policy` | string | Tool approval policy (`always`, `dangerous`, `never`) |
| `db_encryption` | string | Database encryption status |
| `kms_provider` | string | KMS provider name (when configured) |
| `kms_key_id` | string | KMS key identifier (when configured) |
| `kms_fallback` | string | KMS fallback status (when configured) |

---

## lango security migrate-passphrase

Rotate the encryption passphrase. Re-encrypts all stored secrets with a new passphrase. This command requires an interactive terminal.

```
lango security migrate-passphrase
```

!!! warning "Important"
    - Only available when using the `local` security provider
    - Requires an interactive terminal for passphrase input
    - Back up your data directory before running this command
    - If the process is interrupted, data may be corrupted

**Process:**

1. Your current passphrase is verified during bootstrap
2. You are prompted to enter and confirm a new passphrase
3. A new random salt is generated
4. All secrets are decrypted with the old passphrase and re-encrypted with the new one
5. The new salt and passphrase checksum are saved

**Example:**

```bash
$ lango security migrate-passphrase
This process will re-encrypt all your stored secrets with a new passphrase.
Warning: If this process is interrupted, your data may be corrupted.
Ensure you have a backup of your data directory.

Enter NEW passphrase:
Confirm NEW passphrase:
Migrating secrets...
Migration completed successfully!
```

---

## OS Keyring

Manage OS keyring passphrase storage. The OS keyring (macOS Keychain, Linux secret-service, Windows Credential Manager) allows lango to unlock automatically without a keyfile or interactive prompt.

### lango security keyring store

Store the master passphrase in the OS keyring. Requires an interactive terminal. The passphrase is verified against the existing crypto state before storing.

```
lango security keyring store
```

!!! warning "Interactive Only"
    This command requires an interactive terminal and cannot be used in CI/CD pipelines.

**Example:**

```bash
$ lango security keyring store
Enter passphrase to store in keyring: ********
Passphrase stored in OS keyring (macOS Keychain).
```

---

### lango security keyring clear

Remove the master passphrase from the OS keyring.

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
Remove passphrase from OS keyring? [y/N] y
Passphrase removed from OS keyring.

# Non-interactive
$ lango security keyring clear --force
Passphrase removed from OS keyring.
```

---

### lango security keyring status

Show OS keyring availability and stored passphrase status.

```
lango security keyring status [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango security keyring status
OS Keyring Status
  Available:      true
  Backend:        macOS Keychain
  Has Passphrase: true
```

**JSON output fields:**

| Field | Type | Description |
|-------|------|-------------|
| `available` | bool | Whether OS keyring is available |
| `backend` | string | Keyring backend name |
| `error` | string | Error message if unavailable |
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
