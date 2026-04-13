## MODIFIED Requirement: KMS bootstrap fallback description

_Modifies: "KMS bootstrap with graceful fallback" scenario (main spec line 45-49)_

#### Scenario: KMS bootstrap with graceful fallback

- **WHEN** KMS unwrap fails (provider init error, decrypt error, or no matching slot)
- **THEN** the bootstrap SHALL print a warning to stderr
- **AND** fall through to the standard passphrase credential acquisition
- **AND** NOT attempt local decryption of KMS-wrapped ciphertext

_Note: The previous fallback description `(mnemonic → passphrase)` is replaced with `passphrase` only, reflecting the removal of the mnemonic choice during bootstrap._
