## Overview

This is a behavior-locking test-only change.

## Design Decisions

### Transport-level inspection instead of slow timeout simulation

The new test injects a custom `http.RoundTripper` that inspects the outgoing request and returns a synthetic success response. This verifies:

- request method
- request path propagation
- request context deadline

without sleeping for long durations or depending on wall-clock timeout expiration.

### Existing behavior tests stay in place

The success, HTTP error, empty body, and `GetFile` error tests remain unchanged. The new test only tightens the contract around how the download request is issued.
