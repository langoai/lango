## Design Summary

This slice extends the cockpit dead-letter filter bar with:

- `any_match_family`

Supported values:

- `all`
- `retry`
- `manual-retry`
- `dead-letter`

The page keeps the existing interaction model:

- draft edit
- `Enter` apply
- backlog reload
- first-row reset
- detail reload

The cockpit bridge forwards the any-match family value to the existing dead-letter list surface.
