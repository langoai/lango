## Design Summary

This slice extends the cockpit dead-letter filter bar with:

- `latest_status_subtype`

Supported values:

- `all`
- `retry-scheduled`
- `manual-retry-requested`
- `dead-lettered`

The page keeps the existing interaction model:

- draft edit
- `Enter` apply
- backlog reload
- first-row reset
- detail reload

The cockpit bridge forwards the subtype value to the existing dead-letter list surface.
