# cli-prompt-helpers Specification

## Purpose
Define the shared CLI prompt helper contracts for visible and hidden input, confirmation flows, stream seams, and interactive-terminal safety behavior.
## Requirements
### Requirement: Default confirmation wrapper supports deterministic stream seams
The shared `prompt.Confirm(...)` helper SHALL allow its default input and output streams to be replaced in tests without changing runtime confirmation behavior.

#### Scenario: Wrapper uses injected streams under test
- **WHEN** `prompt.Confirm(...)` is exercised in tests with injected default input and output streams
- **THEN** the wrapper SHALL read from the injected input stream
- **AND** it SHALL write the confirmation prompt to the injected output stream

#### Scenario: Wrapper preserves existing confirmation semantics
- **WHEN** `prompt.Confirm(...)` receives `y` or `yes`
- **THEN** it SHALL return approval
- **AND** non-affirmative input SHALL continue to return denial

#### Scenario: Wrapper treats EOF as denial
- **WHEN** `prompt.Confirm(...)` reads EOF before an approval answer is received
- **THEN** it SHALL return `false` with no error

#### Scenario: Wrapper preserves existing runtime defaults
- **WHEN** production code calls `prompt.Confirm(...)` without overriding any seams
- **THEN** the helper SHALL continue to use the process default stdin and stdout streams

### Requirement: Shared line prompt supports deterministic command streams
The shared CLI prompt package SHALL provide a visible line-entry helper that writes a prompt through a supplied output stream and reads one line from a supplied input stream without requiring process-global stdio replacement.

#### Scenario: Shared line prompt uses injected streams
- **WHEN** the visible line-entry helper is exercised in tests with injected input and output streams
- **THEN** it SHALL write the prompt to the injected output stream
- **AND** it SHALL read the entered line from the injected input stream

### Requirement: Shared guarded confirmation supports TTY-only command flows
The shared CLI prompt package SHALL provide a confirmation helper that rejects non-terminal stdin with a caller-supplied guidance error and treats EOF as a clean denial instead of a hard failure.

#### Scenario: Guarded confirmation rejects non-terminal stdin
- **WHEN** the guarded confirmation helper is called with a non-terminal `*os.File` input stream
- **THEN** it SHALL return a caller-supplied error
- **AND** it SHALL NOT write a confirmation prompt

#### Scenario: Guarded confirmation treats EOF as denial
- **WHEN** the guarded confirmation helper reaches EOF before an approval answer is read
- **THEN** it SHALL return `false` with no error

### Requirement: Shared TTY-input guard supports command-specific guidance
The shared CLI prompt package SHALL provide a helper that rejects non-terminal `*os.File` input streams with a caller-supplied guidance error while allowing other reader types to proceed for tests and explicit injected-input flows.

#### Scenario: TTY-input guard rejects non-terminal file input
- **WHEN** the shared TTY-input guard receives a non-terminal `*os.File` input stream
- **THEN** it SHALL return the caller-supplied guidance error

#### Scenario: TTY-input guard allows injected non-file readers
- **WHEN** the shared TTY-input guard receives an injected non-file reader such as `bytes.Buffer`
- **THEN** it SHALL return nil so command tests can continue driving the interaction through explicit streams

### Requirement: Shared interactive-terminal guard supports command-specific guidance
The shared CLI prompt package SHALL provide a helper that fails when the current stdin is not an interactive terminal and returns a caller-supplied guidance error.

#### Scenario: Interactive-terminal guard fails in non-interactive mode
- **WHEN** the shared interactive-terminal guard runs while stdin is not an interactive terminal
- **THEN** it SHALL return the caller-supplied error

#### Scenario: Interactive-terminal guard passes in interactive mode
- **WHEN** the shared interactive-terminal guard runs while stdin is interactive
- **THEN** it SHALL return nil

### Requirement: Shared interactive guard supports explicit input streams
The shared CLI prompt package SHALL provide an interactive guard helper that validates a caller-supplied input stream and returns a caller-supplied guidance error when that stream is a non-interactive terminal file.

#### Scenario: Explicit guard rejects non-terminal file input
- **WHEN** the explicit interactive guard receives a non-terminal `*os.File` input stream
- **THEN** it SHALL return the caller-supplied guidance error

#### Scenario: Explicit guard allows injected readers
- **WHEN** the explicit interactive guard receives an injected non-file reader such as `bytes.Buffer`
- **THEN** it SHALL return nil so command tests and embedded wrappers can drive the subsequent interaction through explicit streams

### Requirement: CLI prompt helpers use shared raw line reader
The shared CLI prompt package SHALL build its visible line-entry prompt helper on top of the shared raw line reader instead of owning a second local line-reader implementation.

#### Scenario: CLI prompt helper delegates raw line reading
- **WHEN** `ReadLineIO(...)` reads from an injected stream
- **THEN** the visible prompt helper SHALL delegate raw line reading to the shared lower-level helper

### Requirement: Shared confirmation wrapper can treat EOF as denial
The shared CLI prompt package SHALL provide a confirmation wrapper that maps EOF to `(false, nil)` while preserving normal approval/denial semantics for explicit input.

#### Scenario: EOF becomes clean denial
- **WHEN** the shared EOF-deny confirmation wrapper reads EOF before an approval answer is received
- **THEN** it SHALL return `false` with no error

### Requirement: Default confirmation wrapper treats EOF as denial
The top-level `prompt.Confirm(...)` wrapper SHALL use the safer EOF-deny confirmation behavior by default.

#### Scenario: Default confirmation wrapper maps EOF to denial
- **WHEN** `prompt.Confirm(...)` reads EOF before an approval answer is received
- **THEN** it SHALL return `false` with no error

### Requirement: Shared hidden passphrase confirmation supports explicit output streams
The shared CLI prompt package SHALL provide a passphrase confirmation helper that writes all visible hidden-input prompt text through a supplied output stream while preserving terminal-hidden password reading.

#### Scenario: Passphrase confirmation uses explicit output
- **WHEN** the explicit-output passphrase confirmation helper prompts for a passphrase and its confirmation
- **THEN** both visible prompt strings SHALL be written through the supplied output stream
- **AND** the helper SHALL return the confirmed passphrase when both hidden inputs match

#### Scenario: Passphrase confirmation mismatch still fails
- **WHEN** the explicit-output passphrase confirmation helper receives different hidden input values
- **THEN** it SHALL return a mismatch error
- **AND** it SHALL NOT return either entered value
