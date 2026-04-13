## MODIFIED Requirement: Recovery is an explicit CLI action

_Replaces: "Recovery during bootstrap" (main spec lines 71-86)_

Mnemonic recovery SHALL NOT be offered as an automatic prompt during bootstrap. Recovery is an explicit user action performed via `lango security recovery restore`. The restore command SHALL load the envelope directly from the filesystem without running the full bootstrap pipeline.

#### Scenario: Bootstrap does not prompt for mnemonic

- **WHEN** bootstrap Phase 4 (AcquireCredential) runs with an envelope containing a mnemonic slot
- **THEN** the mnemonic choice prompt SHALL NOT be shown
- **AND** passphrase acquisition SHALL proceed via the normal priority chain (KMS, keyring, keyfile, interactive, stdin)

#### Scenario: Mnemonic recovery via dedicated command without bootstrap

- **WHEN** the user runs `lango security recovery restore`
- **THEN** the command SHALL load the envelope directly via `security.LoadEnvelopeFile(langoDir)` without invoking the full bootstrap pipeline
- **AND** the user SHALL be prompted for the 24-word mnemonic
- **AND** on success, the user SHALL set a new passphrase via `ChangePassphraseSlot`

#### Scenario: Restore reports clear error when no envelope exists

- **WHEN** the user runs `lango security recovery restore` and no envelope file exists
- **THEN** the command SHALL return an error: `"envelope not found — recovery requires local encryption mode"`

#### Scenario: Non-interactive restore fails gracefully

- **WHEN** `lango security recovery restore` is run in a non-interactive environment
- **THEN** it SHALL return an error requiring an interactive terminal
