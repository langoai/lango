# OS Keyring Integration

## Overview

OS-native keyring integration for secure passphrase storage using macOS Keychain, Linux secret-service (D-Bus), or Windows DPAPI.

## Interface

```go
// Provider abstracts OS keyring operations.
type Provider interface {
    Get(service, key string) (string, error)
    Set(service, key, value string) error
    Delete(service, key string) error
}
```

## Constants

- Service: `"lango"`
- Key: `"master-passphrase"`

## Priority Chain

1. Keyring (if provider set and available)
2. Keyfile (`~/.lango/keyfile`)
3. Interactive terminal prompt
4. stdin pipe

## Availability Detection

`IsAvailable()` performs a write/read/delete probe cycle to verify the OS keyring daemon is accessible. Returns `Status{Available, Backend, Error}`.

## Configuration

```yaml
security:
  keyring:
    enabled: true  # default
```

## CLI Commands

- `lango security keyring status` — show keyring availability
- `lango security keyring store` — store passphrase in keyring
- `lango security keyring clear` — remove passphrase from keyring

## Graceful Fallback

When keyring is unavailable (CI, headless Linux, SSH session), the system silently falls back to keyfile-based passphrase acquisition with no user-visible error.
