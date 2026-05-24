## Why

The economy escrow wrapper contract was tightened so `economy_escrow_create` now requires `milestones` at the wrapper layer. The system prompt documentation still described that tool in looser terms, which leaves the agent-facing prompt surface behind the implemented contract.

## What Changes

- Update `prompts/TOOL_USAGE.md` so `economy_escrow_create` explicitly documents `buyerDid`, `sellerDid`, `amount`, and `milestones` as required inputs.
- Sync `p2p-agent-prompts` coverage for that required-input wording.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `p2p-agent-prompts`: the economy tool usage guidance now matches the enforced required-input contract for economy escrow creation.

## Impact

- Affected prompts: `prompts/TOOL_USAGE.md`
- Affected specs: `openspec/specs/p2p-agent-prompts/spec.md`
