---
title: Configuration Basics
---

# Configuration Basics

Lango stores all configuration in encrypted SQLite profiles. API keys, tokens, and settings are encrypted at rest using AES-256-GCM.

## How Configuration Works

Configuration is stored in `~/.lango/lango.db` as encrypted profiles. Each profile contains a complete set of settings -- provider credentials, channel tokens, feature flags, and more.

!!! info "Encryption"

    All configuration data is encrypted with a passphrase you set on first run. The passphrase is required every time you start Lango or access config commands.

## Guided Setup

For initial setup or quick changes, use the onboard wizard:

```bash
lango onboard
```

This walks you through provider, agent, channel, security, and validation steps. See [Quick Start](quickstart.md) for details.

## Full Editor

For access to every configuration option:

```bash
lango settings
```

This opens an interactive TUI editor where you can browse and modify all settings.

## Config CLI Commands

Manage profiles from the command line:

### List Profiles

```bash
lango config list
```

Shows all profiles with their name, active status, version, and timestamps.

### Create a Profile

```bash
lango config create <name>
```

Creates a new profile with default configuration values.

### Switch Profiles

```bash
lango config use <name>
```

Sets the specified profile as active. The active profile is loaded when you run `lango serve`.

### Delete a Profile

```bash
lango config delete <name>
```

Permanently deletes a profile. Prompts for confirmation unless `--force` is passed.

```bash
lango config delete <name> --force
```

### Import Configuration

```bash
lango config import <file.json> --profile <name>
```

Imports a plaintext JSON configuration file, encrypts it, and stores it as a profile.

!!! warning

    The source JSON file is **deleted after import** for security. Make sure you have a backup if needed.

### Export Configuration

```bash
lango config export <name>
```

Exports a profile as plaintext JSON to stdout. Requires passphrase verification.

!!! warning

    The exported output contains sensitive values (API keys, tokens) in plaintext. Handle with care.

### Validate Configuration

```bash
lango config validate
```

Validates the active profile against required fields and configuration rules. Run this after making changes to catch errors before starting the server.

## Summary

| Task | Command |
|---|---|
| First-time setup | `lango onboard` |
| Edit all settings | `lango settings` |
| List profiles | `lango config list` |
| Create profile | `lango config create <name>` |
| Switch profile | `lango config use <name>` |
| Delete profile | `lango config delete <name>` |
| Import JSON | `lango config import <file>` |
| Export JSON | `lango config export <name>` |
| Validate | `lango config validate` |

## Next Steps

- [Quick Start](quickstart.md) -- Run the onboard wizard
- [CLI Reference](../cli/index.md) -- Full command documentation
- [Security](../security/index.md) -- Encryption and secrets management details
