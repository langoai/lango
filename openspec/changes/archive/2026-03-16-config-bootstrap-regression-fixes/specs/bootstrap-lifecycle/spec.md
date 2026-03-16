## MODIFIED Requirements

### Requirement: Profile loading applies PostLoad normalization
The `phaseLoadProfile` phase SHALL call `config.PostLoad()` exactly once at the end, after all branches (explicit profile, active profile, default profile) have set the config. No branch SHALL return early before PostLoad is applied.

#### Scenario: Explicit profile gets PostLoad applied
- **WHEN** `ForceProfile` is set and the profile is loaded successfully
- **THEN** `PostLoad()` is called on the loaded config before the phase completes

#### Scenario: Active profile gets PostLoad applied
- **WHEN** an active profile exists and is loaded
- **THEN** `PostLoad()` is called on the loaded config before the phase completes

#### Scenario: Default profile gets PostLoad applied
- **WHEN** no active profile exists and a default is created via `handleNoProfile`
- **THEN** `PostLoad()` is called on the created config before the phase completes

#### Scenario: PostLoad failure fails the phase
- **WHEN** `PostLoad()` returns an error on the loaded config
- **THEN** the phase returns that error wrapped with context
