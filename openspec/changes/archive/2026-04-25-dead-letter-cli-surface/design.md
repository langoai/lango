## Design Summary

This slice adds two CLI commands:

- `lango status dead-letters`
- `lango status dead-letter <transaction-receipt-id>`

The CLI reuses the existing dead-letter read model:

- backlog list
- per-transaction detail

First-slice list filters:

- `--query`
- `--adjudication`

Output:

- default `table`
- optional `json`
