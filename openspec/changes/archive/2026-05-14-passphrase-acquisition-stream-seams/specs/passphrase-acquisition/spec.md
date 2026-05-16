## MODIFIED Requirements

### Requirement: Log keyring read errors to stderr
When `passphrase.Acquire()` attempts to read from the OS keyring and receives an error other than `ErrNotFound`, it SHALL write a warning to stderr in the format: `warning: keyring read failed: <error>`. The function SHALL still fall through to the next passphrase source (keyfile, interactive, stdin).

#### Scenario: Keyring returns non-NotFound error
- **WHEN** `KeyringProvider.Get()` returns an error that is not `ErrNotFound`
- **THEN** stderr SHALL contain `warning: keyring read failed: <error detail>`
- **AND** acquisition SHALL continue to the next source

### Requirement: Passphrase acquisition stream seams
The stdin-pipe and keyring-warning branches of passphrase acquisition SHALL be structured so they can be exercised with injected readers and writers in tests, without replacing process-global stdin or stderr.

#### Scenario: Stdin pipe path supports injected reader
- **WHEN** the stdin-pipe branch of acquisition is exercised in tests
- **THEN** the implementation SHALL be able to read from an injected reader instead of requiring `os.Stdin` replacement

#### Scenario: Keyring warning path supports injected writer
- **WHEN** the keyring read branch returns a non-`ErrNotFound` error in tests
- **THEN** the warning path SHALL be capturable via an injected writer instead of requiring `os.Stderr` interception
