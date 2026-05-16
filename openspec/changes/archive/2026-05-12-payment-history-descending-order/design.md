## Overview

This is a test-only contract hardening change for payment history.

## Design Decisions

### Explicit newer record instead of timing dependence

The regression seeds a record with a later `created_at` value using Ent's setter so the sort expectation is deterministic. This avoids flaky timing assumptions between successive inserts.

### Limit is verified on top of sort

The `limit=1` check now verifies that the newest record is the one returned, proving that limit semantics are being applied after the descending sort contract.
