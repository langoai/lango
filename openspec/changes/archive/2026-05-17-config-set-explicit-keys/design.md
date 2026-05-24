# Design: Config Set Explicit Keys

## Approach

Move explicit-key handling into the `configcmd.NewSetCmd` seam so behavior is testable without full bootstrap or encrypted storage.

`NewSetCmd` should receive the loaded profile's explicit-key map together with the config object, mutate a copy of that map after `setConfigPath` succeeds, and pass the updated map to the save callback. The root command wiring should provide `boot.ExplicitKeys` from the single bootstrap result and save with the same profile name.

`onboard` should retain explicit-key metadata returned by `ConfigProfileStore.Load` and pass it through to `Save` when editing an existing profile. When onboarding a missing profile from a preset, it should save `config.PresetExplicitKeys(preset)` so preset-owned context choices remain explicit.

## Rules

- Existing explicit keys must be preserved exactly.
- If the set path matches one of `config.ContextRelatedKeys()`, the saved explicit-key map must include that path.
- If the loaded map is nil and the set path is not context-related, the save callback may receive nil.
- If the loaded map is nil and the set path is context-related, the command must allocate a map containing that path.
- Invalid paths must fail before modifying the explicit-key map or saving.
- Cleanup must still run via defer on success, set-path errors, and save errors.
- Onboard cancellation or incomplete wizard exits must not save or activate profiles.

## Testing

Use RED-first tests in `internal/cli/configcmd` to verify the command seam:

- Unrelated config set preserves an existing explicit disable such as `knowledge.enabled`.
- Setting `knowledge.enabled false` marks that key explicit.
- Invalid paths do not call save and do not mutate the input map.

Add a small root wiring test if needed to ensure `boot.ExplicitKeys` is passed into `NewSetCmd`.

Add onboard tests to verify an existing profile save preserves loaded explicit keys and a new preset-backed profile save includes preset explicit keys.
