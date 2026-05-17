# Fail Doctor Missing Agent Provider

## Why

`lango doctor` can report AI provider configuration as passing when `agent.provider` is set but the `providers` map is empty. That produces a misleading `Configured providers: ` success message even though runtime provider resolution will not have a configured provider entry to use.

## What Changes

- Treat any non-empty `agent.provider` that is absent from the `providers` map as a doctor failure, including when the map is empty.
- Preserve the legacy `GOOGLE_API_KEY` fallback only when neither `agent.provider` nor modern providers are configured.
- Add a focused regression test for the missing agent provider reference.

## Impact

This is a diagnostic correctness change only. It does not change provider runtime resolution or onboarding behavior.
