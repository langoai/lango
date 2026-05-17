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

Create a new configuration profile with default values or a preset template.

```
lango config create <name> [--preset <name>]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `name` | Yes | Name for the new profile |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--preset` | string | `""` | Preset template: `minimal`, `researcher`, `collaborator`, or `full` |

Without `--preset`, the profile is created with Lango's default configuration. Use `lango settings` or `lango onboard` to customize it. With `--preset`, the profile starts from a purpose-built template.

**Example:**

```bash
$ lango config create production
Profile "production" created with default configuration.

$ lango config create research-bot --preset researcher
Profile "research-bot" created from preset "researcher".
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

Delete a configuration profile. Prompts for confirmation unless `--force` is specified. The confirmation prompt is driven through the Cobra command input/output streams so wrappers and test harnesses can capture or drive the interaction via `cmd.OutOrStdout()` and `cmd.InOrStdin()`.

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
The exported JSON contains sensitive values (API keys, tokens) in plaintext. Handle with care and do not commit to version control. Both `lango config export` and `lango config get --output json` emit valid pretty-printed JSON directly on the Cobra command output stream, so wrappers and test harnesses can capture and decode them safely. `lango config get` accepts only `plain` or `json` for `--output` and rejects unknown values before loading the active config.

You can redirect the output to a file:

```bash
$ lango config export default > backup.json
```

---

## lango config get

Read a configuration value by dot-notation path. This is a read-only command and does not require editing the profile.

```
lango config get <dot.path> [--output plain|json]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `dot.path` | Yes | Configuration path such as `agent.provider` or `p2p.enabled` |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output`, `-o` | string | `plain` | Output format (`plain` or `json`) |

The command writes through the Cobra command output stream, and `--output` accepts only `plain` or `json`. Unknown values fail before the active config is loaded.

**Examples:**

```bash
$ lango config get agent.provider
anthropic

$ lango config get p2p.enabled
false

$ lango config get agent --output json
{
  "provider": "anthropic",
  ...
}
```

---

## lango config set

Set a configuration value by dot-notation path. This is a profile-modifying command and therefore requires bootstrap/passphrase verification.

```
lango config set <dot.path> <value>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `dot.path` | Yes | Configuration path such as `agent.provider` or `p2p.enabled` |
| `value` | Yes | New value encoded as text |

**Examples:**

```bash
$ lango config set agent.provider openai
Set agent.provider = openai

$ lango config set p2p.enabled true
Set p2p.enabled = true

$ lango config set providers.openai.type openai
Set providers.openai.type = openai
```

!!! info
    Simple scalar values can be changed here, including scalar fields inside map-backed paths such as `providers.<id>.*` and `mcp.servers.<name>.env.<KEY>`. For complex nested structures and duration-heavy settings, prefer `lango settings`.

!!! warning
    Avoid using `lango config set` examples or shell history for raw secrets. Credential fields may be environment-expanded during profile save. When a sensitive path is set, the success confirmation prints `<redacted>` instead of echoing the provided value.

---

## lango config keys

List available configuration keys derived from the config schema. Optionally filter by a dot-path prefix.

```
lango config keys [prefix]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `prefix` | No | Optional prefix such as `agent` or `p2p.zkp` |

**Examples:**

```bash
$ lango config keys
agent.provider
agent.model
...

$ lango config keys agent
agent.provider
agent.model
agent.temperature
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
