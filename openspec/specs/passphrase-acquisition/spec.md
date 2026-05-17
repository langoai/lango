## Purpose

Capability spec for passphrase-acquisition. See requirements below for scope and behavior contracts.
## Requirements
### Requirement: Passphrase acquisition priority chain
The system SHALL acquire a passphrase using the following priority: (1) hardware keyring (Touch ID / TPM), (2) keyfile at `~/.lango/keyfile`, (3) interactive terminal prompt, (4) stdin pipe. The system SHALL return an error if no source is available.

#### Scenario: Empty stdin pipe is treated as empty passphrase input
- **WHEN** the stdin-pipe passphrase path reaches EOF without reading any passphrase bytes
- **THEN** it SHALL return an empty-input passphrase error instead of a raw read failure

### Requirement: Log keyring read errors to stderr
When `passphrase.Acquire()` attempts to read from the OS keyring and receives an error other than `ErrNotFound`, it SHALL write a warning to stderr in the format: `warning: keyring read failed: <error>`. The function SHALL still fall through to the next passphrase source (keyfile, interactive, stdin).

#### Scenario: Keyring returns non-NotFound error
- **WHEN** `KeyringProvider.Get()` returns an error that is not `ErrNotFound`
- **THEN** stderr SHALL contain `warning: keyring read failed: <error detail>`
- **AND** acquisition SHALL continue to the next source

### Requirement: Passphrase acquisition stream seams

The stdin-pipe and keyring-warning branches of passphrase acquisition SHALL be structured so they can be exercised with injected readers and writers in tests, without replacing process-global stdin or stderr. Public acquisition wrappers SHALL route their process stdio dependencies through package-level seams before delegating to lower-level helpers.

#### Scenario: Public acquire wrapper supports injected stdin seam

- **WHEN** `Acquire(...)` falls through to the stdin-pipe branch with terminal detection disabled
- **THEN** it SHALL read the passphrase from the injected stdin seam
- **AND** it SHALL NOT require replacing process-global `os.Stdin`

#### Scenario: Public acquire wrapper supports injected stderr seam

- **WHEN** `Acquire(...)` observes a keyring read error other than `ErrNotFound`
- **THEN** it SHALL write the warning to the injected stderr seam
- **AND** it SHALL NOT require intercepting process-global `os.Stderr`

#### Scenario: Stdin pipe path supports injected reader

- **WHEN** the stdin-pipe branch of acquisition is exercised in tests
- **THEN** the implementation SHALL be able to read from an injected reader instead of requiring `os.Stdin` replacement

#### Scenario: Keyring warning path supports injected writer

- **WHEN** the keyring read branch returns a non-`ErrNotFound` error in tests
- **THEN** the warning path SHALL be capturable via an injected writer instead of requiring `os.Stderr` interception

#### Scenario: Keyring returns ErrNotFound

- **WHEN** `KeyringProvider.Get()` returns `ErrNotFound`
- **THEN** no warning SHALL be written to stderr
- **AND** acquisition SHALL continue to the next source

### Requirement: Interactive passphrase prompts use acquisition writer
The interactive branch of passphrase acquisition SHALL write visible passphrase prompt text through the writer supplied to the acquisition stream seam.

#### Scenario: Existing passphrase prompt uses injected writer
- **WHEN** `acquireWithIO(...)` falls through to interactive acquisition with creation disabled
- **THEN** the visible `Enter passphrase: ` prompt SHALL be written through the injected writer
- **AND** no prompt text SHALL require process-global stdout capture

#### Scenario: First-run passphrase confirmation uses injected writer
- **WHEN** `acquireWithIO(...)` falls through to interactive acquisition with creation enabled
- **THEN** both visible first-run passphrase prompt strings SHALL be written through the injected writer
- **AND** no prompt text SHALL require process-global stdout capture

#### Scenario: Hidden input behavior is preserved
- **WHEN** interactive acquisition reads the passphrase
- **THEN** hidden input SHALL continue to use the shared terminal password reader path

### Requirement: Keyring provider is nil when no secure hardware is available
The passphrase acquisition flow SHALL receive a nil `KeyringProvider` when the bootstrap determines no secure hardware backend is available (`TierNone`). This effectively disables keyring auto-read, forcing keyfile or interactive/stdin acquisition.

#### Scenario: Nil keyring provider skips keyring step
- **WHEN** `Acquire()` is called with `KeyringProvider` set to nil
- **THEN** the keyring step SHALL be skipped entirely, and acquisition SHALL proceed to keyfile or interactive prompt

#### Scenario: Secure keyring provider attempts read
- **WHEN** `Acquire()` is called with a non-nil `KeyringProvider` (biometric or TPM)
- **THEN** it SHALL attempt to read the passphrase from the secure provider first

### Requirement: Keyfile management
The system SHALL read, write, and securely shred keyfiles with strict 0600 permission enforcement.

#### Scenario: Write keyfile
- **WHEN** a keyfile is written
- **THEN** the file is created with 0600 permissions and parent directories with 0700

#### Scenario: Read keyfile with valid permissions
- **WHEN** a keyfile is read with 0600 permissions
- **THEN** the passphrase is returned with trailing whitespace trimmed

#### Scenario: Read keyfile with invalid permissions
- **WHEN** a keyfile exists with permissions other than 0600
- **THEN** the system returns a permission validation error

#### Scenario: Shred keyfile
- **WHEN** `ShredKeyfile()` is called on an existing keyfile
- **THEN** the file content is overwritten with zero bytes, synced, and removed

#### Scenario: Shred nonexistent keyfile
- **WHEN** `ShredKeyfile()` is called on a nonexistent file
- **THEN** nil is returned without error

### Requirement: Non-interactive passphrase acquisition

The system SHALL provide a `passphrase.AcquireNonInteractive(opts Options)` function that acquires a passphrase only from keyring (Touch ID / TPM) or keyfile, without triggering any interactive terminal prompt or stdin pipe read. This function is used by commands that must work in non-interactive environments (e.g., `lango security status` default path). Warning output emitted by the public wrapper SHALL use the package stderr seam.

#### Scenario: Public non-interactive wrapper supports injected stderr seam

- **WHEN** `AcquireNonInteractive(...)` observes a keyring read error other than `ErrNotFound`
- **THEN** it SHALL write the warning to the injected stderr seam
- **AND** it SHALL NOT require intercepting process-global `os.Stderr`

### Requirement: No recovery credential choice during bootstrap

The bootstrap credential acquisition phase SHALL NOT offer a mnemonic recovery choice. When an envelope contains a mnemonic slot, the bootstrap SHALL proceed with the standard passphrase acquisition chain. Mnemonic recovery is handled exclusively by `lango security recovery restore`.

#### Scenario: Bootstrap with mnemonic slot proceeds normally

- **WHEN** bootstrap Phase 4 runs with an envelope containing a slot of type `KEKSlotMnemonic`
- **THEN** no mnemonic choice prompt SHALL be shown
- **AND** passphrase acquisition SHALL follow the standard priority chain (KMS, keyring, keyfile, interactive, stdin)

#### Scenario: Non-interactive bootstrap unaffected

- **WHEN** bootstrap runs non-interactively (no tty, keyring/keyfile available)
- **THEN** passphrase acquisition SHALL proceed via the normal priority chain
- **AND** no behavior change from the previous non-interactive path

### Requirement: Stdin-pipe acquisition uses shared raw line reader
The non-interactive stdin-pipe passphrase acquisition path SHALL use the shared raw line reader before applying its passphrase-specific trimming and empty-input checks.

#### Scenario: Passphrase stdin path delegates raw line reading
- **WHEN** `ReadStdinPipeFromReader(...)` reads from injected stdin
- **THEN** it SHALL obtain the raw line through the shared lower-level helper
