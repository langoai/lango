## Design Summary

This slice extends the cockpit dead-letter page with a thin filter bar.

The controls are:

- `query`
- `adjudication` (`all`, `release`, `refund`)

Interaction:

- edit draft state
- press `Enter`
- reload filtered backlog
- reset selection to the first row
- reload detail from that row

The page still reuses the existing dead-letter list/detail surfaces and does not introduce live filtering, advanced filter UI, or write controls.
