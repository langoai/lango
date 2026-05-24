## Overview

This is a test-only symmetry fix for RPC wallet lifecycle coverage.

## Design Decisions

### Verify cleanup by request-kind and failure-kind

The new regressions cover both request kinds (`sign_tx`, `sign_msg`) across the major teardown classes:

- normal response
- companion error response
- sender error
- timeout
- context cancellation

That makes the suite reflect the real lifecycle matrix instead of a few representative cases.
