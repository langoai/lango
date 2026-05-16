## ADDED Requirements

### Requirement: Architecture and README inventory docs include current payment-settlement support packages
The public inventory docs SHALL include the shipped payment and settlement support packages that implement approval, direct payment gating, settlement progression, escrow execution, dispute adjudication, and post-adjudication retry/status flows instead of omitting them from the package inventory.

#### Scenario: Payment-settlement package rows stay visible
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal package tree
- **THEN** the architecture inventory SHALL include `finance/`, `paymentapproval/`, `paymentgate/`, `settlementprogression/`, `settlementexecution/`, `partialsettlementexecution/`, `escrowexecution/`, `disputehold/`, `escrowadjudication/`, `escrowrelease/`, `escrowrefund/`, `postadjudicationreplay/`, and `postadjudicationstatus/`
- **AND** the README internal tree SHALL include those same package rows
- **AND** those rows SHALL describe USDC monetary helpers, upfront-payment approval evaluation, direct-payment receipt gating, settlement progression mapping, direct and partial settlement execution, escrow create/fund flow, dispute hold and adjudication, escrow release/refund execution, and post-adjudication retry/status projection truthfully
