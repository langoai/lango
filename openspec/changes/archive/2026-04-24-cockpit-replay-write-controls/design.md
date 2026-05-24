## Design Summary

This slice adds the first write control to the cockpit dead-letter page.

The selected detail pane can now invoke:

- `retry_post_adjudication_execution`

Interaction:

- key binding `r`
- enabled only when `can_retry = true`
- success/failure status message only

The cockpit does not bypass the existing replay path. It only exposes the already landed replay gate and policy gate through an operator-facing UI action.
