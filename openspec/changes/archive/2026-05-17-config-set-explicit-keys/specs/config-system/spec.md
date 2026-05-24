## MODIFIED Requirements

### Requirement: configstore profilePayload
The configstore SHALL wrap Config and ExplicitKeys in a `profilePayload` struct stored inside the encrypted profile. `Save()` SHALL accept explicitKeys parameter. `Load()/LoadActive()` SHALL return explicitKeys alongside Config. Legacy profiles without ExplicitKeys SHALL return nil.

CLI mutation paths that save an already-loaded profile SHALL pass through the loaded `ExplicitKeys` metadata instead of replacing it with nil. Commands that directly set a named context-related key SHALL add that key to `ExplicitKeys` before saving.

#### Scenario: Save with explicitKeys
- **WHEN** `Save(ctx, name, cfg, explicitKeys)` is called
- **THEN** both Config and ExplicitKeys SHALL be encrypted and stored together

#### Scenario: Load legacy profile
- **WHEN** a profile saved before Step 8 is loaded
- **THEN** Config SHALL be returned normally and ExplicitKeys SHALL be nil

#### Scenario: CLI save preserves explicitKeys
- **WHEN** a CLI mutation saves a loaded profile with non-nil `ExplicitKeys`
- **THEN** the saved profile SHALL retain those explicit keys

#### Scenario: Onboard existing profile preserves explicitKeys
- **WHEN** `lango onboard --profile existing` loads and saves an existing profile with non-nil `ExplicitKeys`
- **THEN** the saved profile SHALL retain those explicit keys

#### Scenario: Onboard preset profile saves preset explicitKeys
- **WHEN** `lango onboard --profile research --preset researcher` creates and saves a missing profile
- **THEN** the saved profile SHALL include `PresetExplicitKeys("researcher")`
