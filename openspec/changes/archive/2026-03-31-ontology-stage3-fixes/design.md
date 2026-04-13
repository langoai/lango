## Context

5 fixes from Stage 3 code review. Resolves integration path bugs + security assumption violations.

## Goals / Non-Goals

**Goals:** Fix all 5 issues, add regression tests, pass full build/test.

**Non-Goals:** No new features added.

## Decisions

### D1: TrustScorer interface
Introduce `TrustScorer` interface (GetScore method only) instead of direct `*reputation.Store` reference. Improves testability + dependency separation.

### D2: Post-build wiring
Bridge is carried in `intelligenceValues` and connected to P2P handler at the post-build stage. No additional module dependencies.

### D3: Registry UpdateStatus methods
PromoteType/PromotePredicate use update-only UpdateTypeStatus instead of create-only RegisterType. Reuses DeprecateType pattern.
