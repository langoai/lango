# Design: Sandbox Status Output Format

## Approach

Refactor `lango sandbox status` around a small status snapshot builder. The builder collects the same data currently printed inline: sandbox config, backend selection, active isolator state, platform capabilities, backend availability, Linux-only warning state, and recent audit decisions when bootstrap storage is available.

Rendering then becomes a thin output concern:

- `table`: current multi-section report, preserved as the default.
- `plain`: concise key-value style text for simple shell usage.
- `json`: stable structured payload for scripts and health checks.

## Data Shape

The JSON payload should use snake_case fields and avoid embedding presentation-only strings where booleans or arrays are clearer. Human labels such as the resolved backend string remain useful and can be included alongside raw values.

Recent decisions stay optional. If audit storage is unavailable, the JSON payload should contain an empty `recent_decisions` array and still return success.

## Error Handling

Invalid `--output` values must fail before config/bootstrap loading so scripts get fast feedback without triggering passphrase prompts or database access.

The existing graceful-degradation behavior remains unchanged:

- Prefer `bootLoader` when available.
- Fall back to `cfgLoader` when bootstrap fails.
- Skip recent decisions when audit storage is unavailable.

## Testing

Add RED-first tests that prove:

- `--output json` currently fails or is unavailable, then becomes valid JSON.
- Invalid output values fail before invoking either loader.
- Default table output still includes the established sections.
