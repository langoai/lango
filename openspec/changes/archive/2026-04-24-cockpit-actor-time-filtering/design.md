## Design Summary

This slice extends the cockpit dead-letter filter bar with actor/time inputs.

New controls:

- `manual_replay_actor`
- `dead_lettered_after`
- `dead_lettered_before`

Input model:

- actor remains free-text
- time values remain RFC3339 text inputs

The page keeps the current interaction model:

- draft edit
- `Enter` apply
- backlog reload
- first-row reset
- detail reload

The cockpit bridge forwards the new actor/time values to the existing dead-letter list surface.
