## Design Summary

This slice extends the cockpit dead-letter filter bar with:

- `dead_letter_reason_query`
- `latest_dispatch_reference`

Both controls are text inputs.

The page keeps the existing interaction model:

- draft edit
- `Enter` apply
- backlog reload
- first-row reset
- detail reload

The cockpit bridge forwards both values to the existing dead-letter list surface and omits them when empty.
