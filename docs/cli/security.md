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

Store a new encrypted secret or update an existing one. Prompts for the secret value interactively (input is hidden).

```
lango security secrets set <name>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Name identifier for the secret |

!!! note
    This command requires an interactive terminal. The secret value is read securely without echoing to the screen.

**Example:**

```bash
$ lango security secrets set my-api-key
Enter secret value:
Secret 'my-api-key' stored successfully.
```

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
