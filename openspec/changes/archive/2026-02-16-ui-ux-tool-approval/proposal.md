## Why

The Tool Approval System Enhancement (SafetyLevel, ApprovalPolicy, ExemptTools) has been implemented at the core/application level, but the UI/UX layer still displays the legacy boolean approval model. Users cannot view or change the new `approvalPolicy` setting from the onboarding TUI or CLI, and the README documents the old model.

## What Changes

- Replace `interceptor_approval` (InputBool) with `interceptor_policy` (InputSelect) in the onboard TUI Security form, offering `dangerous`, `all`, `configured`, `none` options
- Add `interceptor_exempt_tools` (InputText, comma-separated) field to the onboard TUI Security form
- Update `UpdateConfigFromForm` to map the new `interceptor_policy` and `interceptor_exempt_tools` fields, removing the legacy `interceptor_approval` case
- Change `security status` CLI output from `ApprovalRequired: enabled/disabled` to `Approval Policy: <policy>`
- Add `approvalPolicy` and `exemptTools` fields to `config.json` example
- Update README Security configuration table with the new fields and deprecation notice for `approvalRequired`

## Capabilities

### New Capabilities

(none — all capabilities already exist)

### Modified Capabilities

- `cli-onboard-security`: Replace boolean approval field with ApprovalPolicy select and ExemptTools text input
- `cli-security-status`: Display ApprovalPolicy string instead of legacy boolean status
- `approval-policy`: Add `exemptTools` and `approvalPolicy` to example config and documentation

## Impact

- `internal/cli/onboard/forms_impl.go` — Security form field changes
- `internal/cli/onboard/state_update.go` — Config mapping changes
- `internal/cli/onboard/forms_impl_test.go` — Test key list update
- `internal/cli/security/status.go` — Status output struct and display
- `config.json` — Example config additions
- `README.md` — Security table update
