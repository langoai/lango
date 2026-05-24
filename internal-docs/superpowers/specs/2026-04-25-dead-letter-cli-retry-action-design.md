# Dead-Letter CLI Retry Action Design

## Purpose / Scope

This design adds the first recovery action to the landed dead-letter CLI surface.

This slice introduces:

- `lango status dead-letter retry <transaction-receipt-id>`

The target is the status CLI dead-letter surface.

This slice directly includes:

- retry command
- retryability precheck
- confirm prompt
- `--yes` bypass
- reuse of the existing replay path

This slice does not directly include:

- polling
- action history
- bulk recovery
- richer follow-up result workflow
- other recovery actions

## Command Surface

The first slice adds one command:

- `lango status dead-letter retry <transaction-receipt-id>`

This keeps the mutation explicit:

- one command
- one transaction target
- one operator action

The slice intentionally avoids:

- hiding mutation behind flags on the detail command
- overloading the existing read-only detail surface

## Precheck Model

The retry command does not jump directly into mutation.

Flow:

1. read the existing dead-letter detail status
2. inspect `can_retry`
3. if `can_retry=false`, fail before mutation
4. if `can_retry=true`, continue

This keeps the CLI aligned with the existing dead-letter status surface and makes the command more understandable to operators.

The final authority still remains the backend replay path. The CLI precheck is an operator-friendly guard, not the canonical gate.

## Confirmation Model

The CLI follows standard interactive behavior:

- default:
  - require a confirm prompt
- `--yes`:
  - skip the prompt

This slice intentionally does not add:

- dry-run mode
- multi-step confirm
- prompt customization

The first slice is just:

- one prompt
- one non-interactive bypass

## Control Reuse

The retry command reuses:

- `retry_post_adjudication_execution`

That means it inherits the existing:

- dead-letter evidence gate
- adjudication gate
- replay policy gate
- evidence append behavior

No new write path or CLI-specific recovery contract is introduced.

## Implementation Shape

Recommended implementation:

- extend `internal/cli/status`
  - add retry subcommand:
    - `dead-letter retry <transaction-receipt-id>`
  - precheck using the existing dead-letter detail read path
  - add interactive confirm prompt
  - add `--yes`
  - invoke the existing replay meta tool bridge
  - support `table` / `json` output for the result

This slice does not add:

- polling loop
- refresh loop
- action history

It is a thin CLI mutation wrapper over the existing control plane.

## Follow-On Inputs

Natural follow-on work after this slice:

1. richer CLI recovery UX
- polling
- structured result output
- richer failure detail

2. more CLI filters
- `any_match_family`
- actor/time
- reason/dispatch

3. broader operator CLI
- grouped summaries
- bulk recovery
