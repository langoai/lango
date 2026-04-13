## MODIFIED Requirement: No recovery credential choice during bootstrap

_Replaces: "Recovery credential choice during bootstrap" (main spec lines 109-124)_

The bootstrap credential acquisition phase SHALL NOT offer a mnemonic recovery choice. When an envelope contains a mnemonic slot, the bootstrap SHALL proceed with the standard passphrase acquisition chain. Mnemonic recovery is handled exclusively by `lango security recovery restore`.

#### Scenario: Bootstrap with mnemonic slot proceeds normally

- **WHEN** bootstrap Phase 4 runs with an envelope containing a slot of type `KEKSlotMnemonic`
- **THEN** no mnemonic choice prompt SHALL be shown
- **AND** passphrase acquisition SHALL follow the standard priority chain (KMS, keyring, keyfile, interactive, stdin)

#### Scenario: Non-interactive bootstrap unaffected

- **WHEN** bootstrap runs non-interactively (no tty, keyring/keyfile available)
- **THEN** passphrase acquisition SHALL proceed via the normal priority chain
- **AND** no behavior change from the previous non-interactive path
