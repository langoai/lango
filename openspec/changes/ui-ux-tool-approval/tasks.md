## 1. Onboard TUI Form Update

- [x] 1.1 Replace `interceptor_approval` (InputBool) with `interceptor_policy` (InputSelect) in `NewSecurityForm` with options ["dangerous", "all", "configured", "none"]
- [x] 1.2 Add `interceptor_exempt_tools` (InputText, comma-separated) field after `interceptor_sensitive_tools`

## 2. Onboard State Update

- [x] 2.1 Add `case "interceptor_policy"` to `UpdateConfigFromForm` mapping to `config.ApprovalPolicy`
- [x] 2.2 Add `case "interceptor_exempt_tools"` with comma-split logic to set `ExemptTools`
- [x] 2.3 Remove legacy `case "interceptor_approval"` from `UpdateConfigFromForm`

## 3. Security Status CLI

- [x] 3.1 Change `statusOutput.ApprovalReq` to `ApprovalPolicy string` with json tag `approval_policy`
- [x] 3.2 Set policy value from `cfg.Security.Interceptor.ApprovalPolicy` (default "dangerous" if empty)
- [x] 3.3 Update text output line from "Approval Required" to "Approval Policy"

## 4. Configuration & Documentation

- [x] 4.1 Add `approvalPolicy` and `exemptTools` fields to `config.json` example
- [x] 4.2 Update README Security table: add `approvalPolicy`, `exemptTools` rows; mark `approvalRequired` as deprecated; change `interceptor.enabled` default to `true`

## 5. Tests

- [x] 5.1 Update `TestNewSecurityForm_AllFields` wantKeys: replace `interceptor_approval` with `interceptor_policy`, add `interceptor_exempt_tools`
