## Why

The brokered-storage change introduced a storage facade, but main-process code still re-entered the database through generic `EntClient()` and `RawDB()` accessors. That left app and CLI code able to issue arbitrary Ent/SQL operations outside broker-owned capability boundaries.

## What Changes

- Replace production `EntClient()`/`RawDB()` consumers with capability-specific storage methods and factories.
- Keep generic raw DB/client wiring internal to bootstrap/test scaffolding only.
- Move CLI history/inspection commands and app wiring to storage-backed readers, factories, and domain bundles.

## Capabilities

### Modified Capabilities
- `brokered-storage`: main-process production code uses capability-specific storage access instead of generic Ent/SQL handles.
- `observability`: alerts route reads from a storage-provided alert reader instead of a raw ent client.
- `cli-learning-inspection`, `cli-librarian-monitoring`, `cli-payment-management`, `cli-workflow-management`, `p2p-reputation-cli`: commands resolve storage through facade capabilities instead of raw ent access.

## Impact

- Affected code: storage facade, app wiring for ontology/observability/p2p, payment/workflow/reputation CLI wiring, learning/inquiry inspection commands.
- Out of scope: removing bootstrap/test-only `WithEntClient` / `WithRawDB` scaffolding or eliminating every internal raw DB use such as FTS bootstrap handles.
