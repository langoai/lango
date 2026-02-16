## Context

The approval-policy capability (ApprovalPolicy type, ExemptTools, needsApproval, migration) is fully implemented at the core/application level. However, the UI/UX layer still uses the legacy boolean `approvalRequired` model:

- Onboard TUI shows `interceptor_approval` as InputBool
- Security status CLI shows "Approval Required: enabled/disabled"
- README documents only the old boolean field
- Example config.json lacks `approvalPolicy` and `exemptTools`

## Goals / Non-Goals

**Goals:**
- Replace the legacy boolean approval form field with an ApprovalPolicy select dropdown
- Add ExemptTools text input to the onboard TUI
- Update security status CLI to display the policy string
- Update documentation (README, example config) to reflect the new model

**Non-Goals:**
- Changing core approval logic (already implemented)
- Removing the deprecated `approvalRequired` field from config struct (kept for migration)
- Adding validation for unknown policy values in the TUI (config loader handles this)

## Decisions

1. **InputSelect for ApprovalPolicy**: Use an InputSelect with four options (`dangerous`, `all`, `configured`, `none`) instead of the boolean toggle. This directly maps to the `config.ApprovalPolicy` type without conversion. Default to `"dangerous"` when the stored value is empty.

2. **Comma-separated ExemptTools**: Reuse the same comma-split pattern as `interceptor_sensitive_tools` for consistency. Both fields use the same parsing logic in `UpdateConfigFromForm`.

3. **Status CLI uses string value**: Replace `boolToStatus()` with direct string output of the policy value. Default to `"dangerous"` if empty, matching the config default.

4. **Legacy field retained in config.json**: Keep `approvalRequired: true` in the example for migration compatibility. Add `approvalPolicy` and `exemptTools` alongside.

## Risks / Trade-offs

- [Migration confusion] Users with legacy configs may see both `approvalRequired` and `approvalPolicy` → Mitigated by migration function in config loader and `(deprecated)` note in README.
- [Empty policy default] If policy field is empty string, TUI defaults to "dangerous" → Consistent with config loader default behavior.
