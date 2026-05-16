## Why

The `p2p_pay` wrapper and the `p2p-payment` spec now require `transaction_receipt_id`, but the agent-facing `TOOL_USAGE.md` prompt still described only `peer_did`, `amount`, and optional `memo`. That leaves the built-in prompt surface behind the actual contract.

## What Changes

- Update `TOOL_USAGE.md` so `p2p_pay` documents `transaction_receipt_id` as required and `submission_receipt_id` as optional.
- Mention the immediate missing-parameter failure when `transaction_receipt_id` is omitted.
- Sync `p2p-agent-prompts` coverage for the updated wording.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `p2p-agent-prompts`: the agent-facing `p2p_pay` tool description now matches the enforced receipt-linked payment contract.

## Impact

- Affected prompts: `prompts/TOOL_USAGE.md`
- Affected specs: `openspec/specs/p2p-agent-prompts/spec.md`
