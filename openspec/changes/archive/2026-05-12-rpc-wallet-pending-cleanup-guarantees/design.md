## Overview

This is a test-only robustness change for RPC wallet lifecycle hygiene.

## Design Decisions

### Cover asymmetry, not just one happy cleanup path

The wallet tests already covered cleanup after one response path and one timeout path. The new regressions extend that coverage to:

- sender error before any response arrives
- context cancellation before any response arrives
- a second request kind (`sign message`)

This closes the most obvious asymmetry without changing runtime code.
