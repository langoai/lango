## Why

The Reputation V2 and runtime-integration code landed ahead of the public documentation. The trust model is now stronger and more explicit in code: composite trust, earned trust, durable negative units, temporary safety signals, and canonical trust-entry states are all runtime-visible.

Public docs and the main `docs-only` OpenSpec spec need to describe that landed behavior instead of treating it as follow-on design work.

## What Changes

- update trust / reputation architecture docs to describe the landed Reputation V2 contract
- update public P2P and economy feature docs to describe canonical trust-entry states and runtime consumption
- narrow the P2P knowledge exchange track follow-on wording to the real remaining gaps
- sync `openspec/specs/docs-only/spec.md`
- archive the completed docs-only workstream

## Impact

- public docs describe the current trust and runtime behavior truthfully
- runtime consumers are documented as reading one canonical trust-entry contract
- docs-only OpenSpec requirements align with the landed behavior instead of stale future-tense wording
