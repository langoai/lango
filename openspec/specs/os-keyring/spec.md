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

`IsAvailable()` performs a write/read/delete probe cycle to verify the OS keyring daemon is accessible. Returns `Status{Available, Backend, SecurityTier, Error}`.

Keyring availability is determined solely by runtime auto-detection — there is no configuration flag.

### Requirement: Status struct describes keyring availability
The `Status` struct SHALL include a `SecurityTier` field indicating the detected hardware security tier alongside existing `Available`, `Backend`, and `Error` fields.

#### Scenario: IsAvailable reports security tier
- **WHEN** `IsAvailable()` is called on a system with biometric hardware
- **THEN** the returned `Status` SHALL have `SecurityTier` set to `TierBiometric`

#### Scenario: IsAvailable on system without secure hardware
- **WHEN** `IsAvailable()` is called on a system without biometric or TPM
- **THEN** the returned `Status` SHALL have `SecurityTier` set to `TierNone`

#### Scenario: Keyring availability on supported OS
- **WHEN** the application starts on a system with an OS keyring daemon
- **THEN** `IsAvailable()` returns `Status{Available: true}` and the keyring is used as the highest-priority passphrase source

#### Scenario: Keyring unavailable in headless environment
- **WHEN** the application starts in a headless environment (CI, Docker, SSH)
- **THEN** `IsAvailable()` returns `Status{Available: false}` and the system silently falls back to keyfile or interactive prompt

## Interactive Keyring Storage Prompt

After a passphrase is acquired interactively (source is `SourceInteractive`) and an OS keyring provider is available, the system SHALL prompt the user to store the passphrase in the OS keyring for future automatic unlock.

#### Scenario: First run with keyring available
- **WHEN** user enters passphrase interactively AND OS keyring is available
- **THEN** system prompts "OS keyring is available. Store passphrase for automatic unlock? [y/N]"

#### Scenario: User accepts keyring storage
- **WHEN** user responds "y" to the keyring storage prompt
- **THEN** system stores the passphrase via `krProvider.Set(Service, KeyMasterPassphrase, pass)`

#### Scenario: User declines keyring storage
- **WHEN** user responds "N" or presses Enter to the keyring storage prompt
- **THEN** system proceeds without storing and does not prompt again until next interactive entry

#### Scenario: Keyring store failure
- **WHEN** user accepts but `krProvider.Set()` returns an error
- **THEN** system prints a warning to stderr and continues startup normally

#### Scenario: Non-interactive passphrase source
- **WHEN** passphrase is acquired from keyring, keyfile, or stdin pipe
- **THEN** system SHALL NOT display the keyring storage prompt

#### Scenario: Keyring unavailable
- **WHEN** OS keyring is not available (headless, CI, Docker)
- **THEN** system SHALL NOT display the keyring storage prompt

## CLI Commands

- `lango security keyring status` — show keyring availability
- `lango security keyring store` — store passphrase in keyring
- `lango security keyring clear` — remove passphrase from keyring

## Graceful Fallback

When keyring is unavailable (CI, headless Linux, SSH session), the system silently falls back to keyfile-based passphrase acquisition with no user-visible error.
