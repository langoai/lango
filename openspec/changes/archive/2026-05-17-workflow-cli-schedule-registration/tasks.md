## 1. Implementation

- [x] Update `lango workflow run --schedule` to register an enabled cron job via bootstrap cron storage.
- [x] Build a deterministic scheduled-workflow automation prompt that references the absolute workflow file path.
- [x] Reject invalid cron schedules before registration using the runtime scheduler parser.
- [x] Preserve validation-first behavior and Cobra output routing.

## 2. Tests

- [x] Replace the not-implemented schedule test with coverage for cron job creation, output text, prompt contents, and bootstrap failure behavior.
- [x] Add negative coverage proving invalid cron schedules are not persisted.
- [x] Run focused workflow CLI tests.

## 3. Docs and Specs

- [x] Update `docs/cli/automation.md` to describe cron-backed scheduled workflow registration.
- [x] Validate the OpenSpec change and sync/archive after implementation.
