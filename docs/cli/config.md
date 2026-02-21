# Config Management

Lango stores all configuration in encrypted profiles within `~/.lango/lango.db`. The `config` command group manages these profiles.

```
lango config <subcommand>
```

---

## lango config list

List all configuration profiles with their status.

```
lango config list
```

**Output columns:**

| Column | Description |
|--------|-------------|
| NAME | Profile name |
| ACTIVE | `*` if this is the active profile |
| VERSION | Configuration schema version |
| CREATED | Creation timestamp |
| UPDATED | Last modification timestamp |

**Example:**

```bash
$ lango config list
NAME       ACTIVE  VERSION  CREATED              UPDATED
default    *       1        2026-01-15 10:00:00  2026-02-20 14:30:00
staging            1        2026-02-01 09:00:00  2026-02-01 09:00:00
```

---

## lango config create

Create a new configuration profile with default values.

```
lango config create <name>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Name for the new profile |

The profile is created with Lango's default configuration. Use `lango settings` or `lango onboard` to customize it.

**Example:**

```bash
$ lango config create production
Profile "production" created with default configuration.
```

!!! warning
    The command fails if a profile with the same name already exists.

---

## lango config use

Switch the active profile. All subsequent commands will use the selected profile's configuration.

```
lango config use <name>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Name of the profile to activate |

**Example:**

```bash
$ lango config use staging
Switched to profile "staging".
```

---

## lango config delete

Delete a configuration profile. Prompts for confirmation unless `--force` is specified.

```
lango config delete <name> [--force]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Name of the profile to delete |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force`, `-f` | bool | `false` | Skip confirmation prompt |

**Examples:**

```bash
# Interactive confirmation
$ lango config delete staging
Delete profile "staging"? This cannot be undone. [y/N]: y
Profile "staging" deleted.

# Non-interactive
$ lango config delete staging --force
Profile "staging" deleted.
```

---

## lango config import

Import a plaintext JSON configuration file into an encrypted profile. The source file is deleted after successful import for security.

```
lango config import <file> [--profile <name>]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `file` | Yes | Path to the JSON configuration file |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--profile` | string | `default` | Name for the imported profile |

**Example:**

```bash
$ lango config import ./config.json --profile production
Imported "./config.json" as profile "production" (now active).
Source file deleted for security.
```

!!! warning "Source File Deleted"
    The source JSON file is automatically deleted after a successful import to prevent sensitive values (API keys, tokens) from remaining on disk in plaintext.

---

## lango config export

Export an encrypted profile as plaintext JSON to stdout. Requires passphrase verification (handled by the bootstrap process).

```
lango config export <name>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Name of the profile to export |

**Example:**

```bash
$ lango config export default
WARNING: exported configuration contains sensitive values in plaintext.
{
  "server": {
    "host": "localhost",
    "port": 18789
  },
  ...
}
```

!!! danger "Security Notice"
    The exported JSON contains sensitive values (API keys, tokens) in plaintext. Handle with care and do not commit to version control.

You can redirect the output to a file:

```bash
$ lango config export default > backup.json
```

---

## lango config validate

Validate the active configuration profile against Lango's schema rules.

```
lango config validate
```

**Example:**

```bash
$ lango config validate
Profile "default" configuration is valid.
```

If validation fails, the command prints the specific errors and exits with a non-zero status code.
