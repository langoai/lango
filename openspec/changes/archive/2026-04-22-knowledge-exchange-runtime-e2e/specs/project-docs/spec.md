## ADDED Requirements

### Requirement: Architecture landing page links the knowledge exchange runtime design slice
The `docs/architecture/index.md` page SHALL include a quick link to `knowledge-exchange-runtime.md` and a short summary that frames it as the first transaction-oriented runtime control-plane design slice for `knowledge exchange v1`, centered on transaction receipt and submission receipt with explicit current limits.

#### Scenario: Knowledge exchange runtime appears in architecture landing page
- **WHEN** a user reads `docs/architecture/index.md`
- **THEN** they SHALL find the Knowledge Exchange Runtime entry linking to `knowledge-exchange-runtime.md`
- **AND** the entry SHALL describe the design slice as transaction-oriented and bounded by current limits

