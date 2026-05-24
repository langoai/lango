## Design Summary

This slice adds:

- `lango status dead-letter retry <transaction-receipt-id>`

Flow:

1. read existing dead-letter detail status
2. check `can_retry`
3. if retryable, ask for confirmation
4. if `--yes`, skip confirmation
5. invoke the existing retry control path

The slice intentionally keeps result handling simple:

- plain success output
- optional `json`
- no polling
